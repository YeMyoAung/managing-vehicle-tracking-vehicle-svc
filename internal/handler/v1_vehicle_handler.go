package handler

import (
    "errors"
    "fmt"
    "log"
    "net/http"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/goccy/go-json"
    "github.com/yemyoaung/managing-vehicle-tracking-common"
    "github.com/yemyoaung/managing-vehicle-tracking-models"
    "github.com/yemyoaung/managing-vehicle-tracking-vehicle-svc/internal/services"
)

var (
    ErrMethodNotAllowed = errors.New("method was not allowed")
    ErrNotFound         = errors.New("not found")
    ErrInvalidRequest   = errors.New("invalid request")
)

type V1TrackingHandler struct {
    vehicleService services.VehicleService
    validate       *validator.Validate
}

func NewV1VehicleHandler(vehicleService services.VehicleService, validate *validator.Validate) *V1TrackingHandler {
    return &V1TrackingHandler{vehicleService: vehicleService, validate: validate}
}

func (h *V1TrackingHandler) methodWasNotAllowed(w http.ResponseWriter) {
    common.HandleError(http.StatusMethodNotAllowed, w, ErrMethodNotAllowed)
}

func (h *V1TrackingHandler) HandleCreateAndFindVehicle(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet && r.Method != http.MethodPost {
        h.methodWasNotAllowed(w)
        return
    }
    if r.Method == http.MethodPost {
        h.CreateVehicle(w, r)
        return
    }
    h.FindVehicles(w, r)
}

func (h *V1TrackingHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
    var req services.VehicleRequest
    if body, ok := r.Context().Value(common.Body).([]byte); ok {
        if err := json.Unmarshal(body, &req); err != nil {
            common.HandleError(http.StatusUnprocessableEntity, w, err)
            return
        }
    }

    if err := h.validate.Struct(&req); err != nil {
        common.HandleError(http.StatusUnprocessableEntity, w, err)
        return
    }

    // if we need to use the user, we can do it like this
    // user, ok := r.Context().Value(common.UserContextKey).(*middlewares.AuthUser)
    // 
    // if ok {
    //     log.Println("User: ", user)
    // }
    vehicle, err := h.vehicleService.CreateVehicle(r.Context(), &req)
    if err != nil {
        common.HandleError(http.StatusUnprocessableEntity, w, err)
        return
    }

    if err = json.NewEncoder(w).Encode(
        common.DefaultSuccessResponse(
            vehicle,
            "successfully created vehicle",
        ),
    ); err != nil {
        log.Printf("Failed to encode response: %v", err)
    }
}

func (h *V1TrackingHandler) FindVehicles(w http.ResponseWriter, r *http.Request) {
    vehicles, err := h.vehicleService.FindVehicles(r.Context(), r.URL.Query())
    if err != nil {
        common.HandleError(http.StatusBadRequest, w, err)
        return
    }

    if len(vehicles) == 0 {
        common.HandleError(http.StatusNotFound, w, ErrNotFound)
        return
    }

    if err = json.NewEncoder(w).Encode(common.DefaultSuccessResponse(vehicles, "successfully fetched vehicles"));
        err != nil {
        log.Printf("Failed to encode response: %v", err)
    }
}

func (h *V1TrackingHandler) FindVehicleByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        h.methodWasNotAllowed(w)
        return
    }

    path := r.URL.Path
    segments := strings.Split(path, "/")

    // path is "/api/v1/vehicles/:id", the ID should be in the fifth segment
    if len(segments) < 5 {
        http.NotFound(w, r)
        return
    }

    vehicle, err := h.vehicleService.GetVehicleByID(r.Context(), segments[4])

    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        if err := json.NewEncoder(w).Encode(common.DefaultErrorResponse(err)); err != nil {
            log.Printf("Failed to encode response: %v", err)
        }
        return
    }

    err = json.NewEncoder(w).Encode(
        common.DefaultSuccessResponse(
            vehicle,
            fmt.Sprintf("successfully fetched vehicle with ID: %s", segments[4]),
        ),
    )

    if err != nil {
        log.Printf("Failed to encode response: %v", err)
    }

}

func (h *V1TrackingHandler) PublishTrackingData(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        h.methodWasNotAllowed(w)
        return
    }

    var req models.TrackingDataRequest

    if body, ok := r.Context().Value(common.Body).([]byte); ok {
        if err := json.Unmarshal(body, &req); err != nil {
            common.HandleError(http.StatusUnprocessableEntity, w, err)
            return
        }
    } else {
        common.HandleError(http.StatusBadRequest, w, ErrInvalidRequest)
    }

    if err := h.validate.Struct(&req); err != nil {
        common.HandleError(http.StatusUnprocessableEntity, w, err)
        return
    }

    // if we need to use the user, we can do it like this
    // user, ok := r.Context().Value(common.UserContextKey).(*middlewares.AuthUser)
    // 
    // if ok {
    //     log.Println("User: ", user)
    // }
    err := h.vehicleService.PublishTrackingData(r.Context(), &req)

    if err != nil {
        common.HandleError(http.StatusUnprocessableEntity, w, err)
        return
    }

    if err = json.NewEncoder(w).Encode(common.DefaultSuccessResponse(nil, "successfully published tracking data"));
        err != nil {
        log.Printf("Failed to encode response: %v", err)
    }

}
