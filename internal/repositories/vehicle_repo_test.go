package repositories

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "strings"
    "testing"

    "github.com/yemyoaung/managing-vehicle-tracking-models"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

const (
    connStr = "mongodb://yoma_fleet:YomaFleet!123@localhost:27017"
)

func getVehicleRepo() (*mongo.Client, *MongoVehicleRepository, error) {
    // we can also use mock database for testing
    // but for now we will use real database to make sure everything is working fine
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connStr))
    if err != nil {
        log.Fatal("Database connection failed:", err)
    }

    repo, err := NewMongoVehicleRepository(context.Background(), client.Database("vehicles"))

    if err != nil {
        return nil, nil, err
    }

    return client, repo, nil
}

var VehicleStatuses = []models.VehicleStatus{
    models.VehicleStatusActive,
    models.VehicleStatusInactive,
    models.VehicleStatusRepair,
    models.VehicleStatusSold,
    models.VehicleStatusRented,
}

func getRandomVehicle() *models.Vehicle {
    vehicle := models.NewVehicle().SetVehicleName(
        fmt.Sprintf(
            "Vehicle %d",
            rand.Int(),
        ),
    ).SetVehicleModel(fmt.Sprintf("Model %d", rand.Int())).
        SetLicenseNumber(fmt.Sprintf("License %d", rand.Int())).
        SetVehicleStatus(VehicleStatuses[rand.Intn(len(VehicleStatuses))]).
        SetMileage(rand.Float64() * 1000)
    if err := vehicle.Build(); err != nil {
        return nil
    }
    return vehicle
}

func TestMongoVehicleRepository_CreateVehicle(t *testing.T) {
    client, repo, err := getVehicleRepo()

    if err != nil {
        t.Fatal(err)
    }

    defer func(client *mongo.Client, ctx context.Context) {
        err := client.Disconnect(ctx)
        if err != nil {
            log.Println("Failed to disconnect from database")
        }
    }(client, context.Background())

    vehicle := getRandomVehicle()

    err = repo.CreateVehicle(context.Background(), vehicle)

    if err != nil {
        t.Fatal(err)
    }

    if vehicle.ID.IsZero() {
        t.Fatal("ID should not be zero")
    }

    err = repo.CreateVehicle(context.Background(), vehicle)

    if err == nil {
        t.Fatal("Error should not be nil")
    }
}

func TestMongoVehicleRepository_FindVehicleByID(t *testing.T) {

    client, repo, err := getVehicleRepo()

    if err != nil {
        t.Fatal(err)
    }

    defer func(client *mongo.Client, ctx context.Context) {
        err := client.Disconnect(ctx)
        if err != nil {
            log.Println("Failed to disconnect from database")
        }
    }(client, context.Background())

    vehicle := getRandomVehicle()

    if err != nil {
        t.Fatal(err)
    }

    err = repo.CreateVehicle(context.Background(), vehicle)

    if err != nil {
        t.Fatal(err)
    }

    var dbVehicle models.Vehicle

    err = repo.FindVehicleByID(context.Background(), vehicle.ID.Hex(), &dbVehicle)

    if err != nil {
        t.Fatal(err)
        return
    }

    if err = dbVehicle.Validate(); err != nil {
        t.Fatal(err)
    }

    if dbVehicle.LicenseNumber != dbVehicle.LicenseNumber {
        t.Fatal("License should be equal")
    }

    var dbVehicle2 models.Vehicle

    err = repo.FindVehicleByID(context.Background(), "6734c2a5eb0eff570b970eb1", &dbVehicle2)

    if err == nil {
        t.Fatal("Error should not be nil")
    }

    if dbVehicle2.Check() == nil {
        t.Fatal("Vehicle should not be valid")
    }
}

func TestMongoVehicleRepository_FindVehicles(t *testing.T) {
    client, repo, err := getVehicleRepo()

    if err != nil {
        t.Fatal(err)
    }

    defer func(client *mongo.Client, ctx context.Context) {
        err := client.Disconnect(ctx)
        if err != nil {
            log.Println("Failed to disconnect from database")
        }
    }(client, context.Background())

    for i := 0; i < 10; i++ {
        vehicle := getRandomVehicle()
        if err := repo.CreateVehicle(context.Background(), vehicle); err != nil {
            t.Fatal(err)
        }
    }

    for i := 1; i <= 5; i++ {
        vehicles, err := repo.FindVehicles(
            context.Background(), &VehicleFilter{
                Page:     i,
                PageSize: 2,
            },
        )
        if err != nil {
            t.Fatal(err)
        }
        if len(vehicles) != 2 {
            t.Fatal("Should return 2 vehicles")
        }
    }

    vehicles, err := repo.FindVehicles(
        context.Background(), &VehicleFilter{
            Page:      1,
            PageSize:  10,
            SortField: "vehicle_name",
        },
    )

    if err != nil {
        t.Fatal(err)
    }

    if len(vehicles) != 10 {
        t.Fatal("Should return 10 vehicles")
    }

    for i := 0; i < len(vehicles)-1; i++ {
        if vehicles[i].VehicleName > vehicles[i+1].VehicleName {
            t.Fatal("Vehicles should be sorted")
        }
    }

    vehicles, err = repo.FindVehicles(
        context.Background(), &VehicleFilter{
            Page:      1,
            PageSize:  10,
            SortField: "vehicle_name",
            SortOrder: "desc",
        },
    )

    if err != nil {
        t.Fatal(err)
    }

    if len(vehicles) != 10 {
        t.Fatal("Should return 10 vehicles")
    }

    for i := 0; i < len(vehicles)-1; i++ {
        if vehicles[i].VehicleName < vehicles[i+1].VehicleName {
            t.Fatal("Vehicles should be sorted")
        }
    }

    for i := 0; i < 10; i++ {
        vehicle := getRandomVehicle()
        vehicle.SetVehicleName(fmt.Sprintf("Testing %d", i))
        if err := repo.CreateVehicle(context.Background(), vehicle); err != nil {
            t.Fatal(err)
        }
    }

    vehicles, err = repo.FindVehicles(
        context.Background(), &VehicleFilter{
            Page:        1,
            PageSize:    10,
            VehicleName: "Test",
        },
    )

    if err != nil {
        t.Fatal(err)
    }

    if len(vehicles) != 10 {
        t.Fatal("Should return 10 vehicles")
    }

    for _, vehicle := range vehicles {
        if !strings.HasPrefix(vehicle.VehicleName, "Test") {
            t.Fatal("Vehicle name should be Test")
        }
    }

    vehicles, err = repo.FindVehicles(
        context.Background(), &VehicleFilter{
            Page:          1,
            PageSize:      10,
            LicenseNumber: vehicles[0].LicenseNumber,
        },
    )

    if err != nil {
        t.Fatal(err)
    }

    if len(vehicles) != 1 {
        t.Fatal("Should return 1 vehicle")
    }

    if vehicles[0].LicenseNumber != vehicles[0].LicenseNumber {
        t.Fatal("License number should be equal")
    }

}

func TestMongoVehicleRepository_UpdateVehicleMileAge(t *testing.T) {
    client, repo, err := getVehicleRepo()

    if err != nil {
        t.Fatal(err)
    }

    defer func(client *mongo.Client, ctx context.Context) {
        err := client.Disconnect(ctx)
        if err != nil {
            log.Println("Failed to disconnect from database")
        }
    }(client, context.Background())

    vehicle := getRandomVehicle()

    err = repo.CreateVehicle(context.Background(), vehicle)

    if err != nil {
        t.Fatal(err)
    }

    mileAge := rand.Float64() * 1000

    err = repo.TrackingVehicle(context.Background(), vehicle.ID.Hex(), mileAge, models.VehicleStatusActive)

    if err != nil {
        t.Fatal(err)
    }

    var dbVehicle models.Vehicle

    err = repo.FindVehicleByID(context.Background(), vehicle.ID.Hex(), &dbVehicle)

    if err != nil {
        t.Fatal(err)
    }

    if dbVehicle.Mileage != mileAge {
        t.Fatal("Mileage should be equal")
    }
}
