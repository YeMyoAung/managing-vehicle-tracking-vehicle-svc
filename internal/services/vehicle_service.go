package services

import (
    "context"
    "net/url"
    "strconv"

    "github.com/goccy/go-json"
    "github.com/yemyoaung/managing-vehicle-tracking-models"
    "github.com/yemyoaung/managing-vehicle-tracking-vehicle-svc/internal/repositories"
)

type VehicleRequest struct {
    VehicleName   string               `json:"vehicle_name"  validate:"required"`
    VehicleModel  string               `json:"vehicle_model" validate:"required"`
    VehicleStatus models.VehicleStatus `json:"vehicle_status" validate:"required"`
    Mileage       float64              `json:"mileage" validate:"required"`
    LicenseNumber string               `json:"license_number" validate:"required"`
}

func (v *VehicleRequest) Validate() error {
    if err := v.VehicleStatus.Valid(); err != nil {
        return err
    }
    return nil
}

type VehicleService interface {
    CreateVehicle(ctx context.Context, req *VehicleRequest) (*models.Vehicle, error)
    TrackingVehicle(ctx context.Context, id string, mileAge float64, status models.VehicleStatus) error
    FindVehicles(ctx context.Context, query url.Values) ([]*models.Vehicle, error)
    GetVehicleByID(ctx context.Context, id string) (*models.Vehicle, error)
    PublishTrackingData(ctx context.Context, req *models.TrackingDataRequest) error
}

type MongoVehicleService struct {
    vehicleRepo  repositories.VehicleRepository
    trackingRepo repositories.TrackingRepository
}

func NewMongoVehicleService(
    vehicleRepo repositories.VehicleRepository,
    trackingRepo repositories.TrackingRepository,
) *MongoVehicleService {
    return &MongoVehicleService{
        vehicleRepo:  vehicleRepo,
        trackingRepo: trackingRepo,
    }
}

func (s *MongoVehicleService) CreateVehicle(ctx context.Context, req *VehicleRequest) (*models.Vehicle, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }
    vehicle := models.NewVehicle().
        SetVehicleName(req.VehicleName).
        SetVehicleModel(req.VehicleModel).
        SetVehicleStatus(req.VehicleStatus).
        SetMileage(req.Mileage).
        SetLicenseNumber(req.LicenseNumber)
    err := s.vehicleRepo.CreateVehicle(ctx, vehicle)
    if err != nil {
        return nil, err
    }
    return vehicle, nil
}

func (s *MongoVehicleService) TrackingVehicle(
    ctx context.Context,
    id string,
    mileAge float64,
    status models.VehicleStatus,
) error {
    return s.vehicleRepo.TrackingVehicle(ctx, id, mileAge, status)
}

func (s *MongoVehicleService) FindVehicles(ctx context.Context, query url.Values) ([]*models.Vehicle, error) {
    // by converting url.Values to map[string]any and unmarshalling it to VehicleFilter,
    // we can ignore unsupported query parameters
    data := map[string]any{}
    for key, value := range query {
        if key == "page" || key == "limit" {
            converted, err := strconv.Atoi(value[0])
            if err != nil {
                return nil, err
            }
            data[key] = converted
            continue
        }
        if key == "mileage" {
            converted, err := strconv.ParseFloat(value[0], 64)
            if err != nil {
                return nil, err
            }
            data[key] = converted
            continue
        }
        data[key] = value[0]
    }

    buf, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }

    var filter repositories.VehicleFilter
    if err := json.Unmarshal(buf, &filter); err != nil {
        return nil, err
    }

    return s.vehicleRepo.FindVehicles(ctx, &filter)
}

func (s *MongoVehicleService) GetVehicleByID(ctx context.Context, id string) (
    *models.Vehicle,
    error,
) {
    var vehicle models.Vehicle
    err := s.vehicleRepo.FindVehicleByID(ctx, id, &vehicle)
    if err != nil {
        return nil, err
    }
    return &vehicle, nil
}

func (s *MongoVehicleService) PublishTrackingData(
    ctx context.Context,
    req *models.TrackingDataRequest,
) error {
    if err := req.Validate(); err != nil {
        return err
    }
    buf, err := json.Marshal(req)
    if err != nil {
        return err
    }
    return s.trackingRepo.PublishTrackingData(ctx, buf)
}
