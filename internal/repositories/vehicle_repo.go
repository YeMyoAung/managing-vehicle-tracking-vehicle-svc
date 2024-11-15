package repositories

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/yemyoaung/managing-vehicle-tracking-models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type VehicleFilter struct {
    Page          int                  `json:"page"`
    PageSize      int                  `json:"limit"`
    SortField     string               `json:"sort_by"`
    SortOrder     string               `json:"sort_order"`
    ID            string               `json:"_id"`
    VehicleName   string               `json:"vehicle_name"`
    VehicleModel  string               `json:"vehicle_model"`
    LicenseNumber string               `json:"license_number"`
    VehicleStatus models.VehicleStatus `json:"vehicle_status"`
    Mileage       float64              `json:"mileage"`
    id            primitive.ObjectID
}

func (v *VehicleFilter) ObjectID() primitive.ObjectID {
    return v.id
}

func (v *VehicleFilter) Build() error {
    if v.Page == 0 {
        v.Page = 1
    }
    if v.PageSize == 0 {
        v.PageSize = 10
    }
    if v.PageSize > 100 {
        v.PageSize = 100
    }
    if v.SortField == "" {
        v.SortField = "created_at"
    }
    if v.SortOrder == "" {
        v.SortOrder = "asc"
    }
    if v.VehicleStatus != "" {
        if err := v.VehicleStatus.Valid(); err != nil {
            return err
        }
    }
    if v.ID != "" {
        objectID, err := primitive.ObjectIDFromHex(v.ID)
        if err != nil {
            return err
        }
        v.id = objectID
    }
    return nil
}

type VehicleRepository interface {
    CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error
    TrackingVehicle(ctx context.Context, id string, mileAge float64, status models.VehicleStatus) error
    FindVehicles(
        ctx context.Context,
        filter *VehicleFilter,
    ) ([]*models.Vehicle, error)
    FindVehicleByID(ctx context.Context, id string, vehicle *models.Vehicle) error
}

type MongoVehicleRepository struct {
    collection *mongo.Collection
}

func NewMongoVehicleRepository(ctx context.Context, db *mongo.Database) (*MongoVehicleRepository, error) {
    vehiclesCollection := db.Collection("vehicles")

    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    indexModel := mongo.IndexModel{
        Keys:    bson.M{"license_number": 1},
        Options: options.Index().SetUnique(true),
    }

    _, err := vehiclesCollection.Indexes().CreateOne(ctx, indexModel)
    if err != nil {
        return nil, err
    }
    return &MongoVehicleRepository{
        collection: vehiclesCollection,
    }, nil
}

func (repo *MongoVehicleRepository) CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error {
    if err := vehicle.Build(); err != nil {
        return err
    }
    result, err := repo.collection.InsertOne(ctx, vehicle)
    if err != nil {
        return err
    }
    vehicle.ID = result.InsertedID.(primitive.ObjectID)
    return nil
}

func (repo *MongoVehicleRepository) TrackingVehicle(
    ctx context.Context,
    id string,
    mileAge float64,
    status models.VehicleStatus,
) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return err
    }
    updateResult, err := repo.collection.UpdateByID(
        ctx,
        objectID,
        bson.M{
            // should use $inc for incrementing mileage, but for the sake of example, we use $set
            "$set": bson.M{
                "mileage":        mileAge,
                "vehicle_status": status,
            },
            "$currentDate": bson.M{"updated_at": true},
        },
    )
    if err != nil {
        return err
    }
    if updateResult.MatchedCount == 0 {
        return fmt.Errorf("vehicle not found")
    }
    if updateResult.UpsertedID != nil {
        if updateResult.UpsertedID.(primitive.ObjectID) != objectID {
            return fmt.Errorf("vehicle not found")
        }
    }
    return nil
}

func (repo *MongoVehicleRepository) FindVehicles(
    ctx context.Context,
    filter *VehicleFilter,
) ([]*models.Vehicle, error) {
    var vehicles []*models.Vehicle

    bsonMFilter := bson.M{}
    findOptions := options.Find()

    if filter != nil {
        if err := filter.Build(); err != nil {
            return nil, err
        }
        if filter.ID != "" {
            bsonMFilter["_id"] = filter.ObjectID()
        }
        if filter.VehicleName != "" {
            bsonMFilter["vehicle_name"] = bson.M{"$regex": fmt.Sprintf("^%s", filter.VehicleName), "$options": "i"}
        }
        if filter.VehicleModel != "" {
            bsonMFilter["vehicle_model"] = bson.M{"$regex": fmt.Sprintf("^%s", filter.VehicleModel), "$options": "i"}
        }
        if filter.VehicleStatus != "" {
            bsonMFilter["vehicle_status"] = filter.VehicleStatus
        }
        if filter.Mileage != 0 {
            bsonMFilter["mileage"] = bson.M{"$gte": filter.Mileage}
        }
        if filter.LicenseNumber != "" {
            bsonMFilter["license_number"] = bson.M{"$regex": fmt.Sprintf("^%s", filter.LicenseNumber), "$options": "i"}
        }

        if filter.SortField != "" {
            order := 1
            if filter.SortOrder == "desc" {
                order = -1
            }
            findOptions.SetSort(bson.D{{Key: filter.SortField, Value: order}})
        }

        findOptions.SetSkip(int64((filter.Page - 1) * filter.PageSize))
        findOptions.SetLimit(int64(filter.PageSize))
    }

    cursor, err := repo.collection.Find(ctx, bsonMFilter, findOptions)
    if err != nil {
        return nil, err
    }
    defer func(cursor *mongo.Cursor, ctx context.Context) {
        err := cursor.Close(ctx)
        if err != nil {
            log.Println("Failed to close cursor", err)
        }
    }(cursor, ctx)

    for cursor.Next(ctx) {
        var vehicle models.Vehicle
        if err := cursor.Decode(&vehicle); err != nil {
            return nil, err
        }
        vehicles = append(vehicles, &vehicle)
    }

    return vehicles, nil
}

func (repo *MongoVehicleRepository) FindVehicleByID(
    ctx context.Context,
    id string,
    vehicle *models.Vehicle,
) error {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return err
    }
    err = repo.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(vehicle)
    if err != nil {
        return err
    }
    if err := vehicle.Check(); err != nil {
        return err
    }
    return nil
}
