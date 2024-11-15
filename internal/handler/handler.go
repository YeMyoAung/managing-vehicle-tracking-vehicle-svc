package handler

import "net/http"

// VehicleHandler is an interface for handling vehicle related requests
type VehicleHandler interface {
    HandleCreateAndFindVehicle(w http.ResponseWriter, r *http.Request)
    FindVehicleByID(w http.ResponseWriter, r *http.Request)
    PublishTrackingData(w http.ResponseWriter, r *http.Request)
}
