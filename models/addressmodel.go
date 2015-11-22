package Models

import (
	"gopkg.in/mgo.v2/bson"
)

type (  
    // User represents the structure of our resource
    TripRequest struct {
        StartingLocation string `json:"starting_from_location_id" bson:"starting_from_location_id"`
		LocationIds[] string `json:"location_ids" bson:"location_ids"`
    }
)

type (  
    // User represents the structure of our resource
    TripResponse struct {
    	Id bson.ObjectId `json:"id" bson:"_id"`
        Status string `json:"status" bson:"status"`
		StartingLocation string `json:"starting_from_location_id" bson:"starting_from_location_id"`
		BestRoute[] string `json:"best_route_location_ids" bson:"best_route_location_ids"`
		TotalCost int `json:"total_uber_costs" bson:"total_uber_costs"`
		TotalDuration int `json:"total_uber_duration" bson:"total_uber_duration"`
		TotalDistance float64 `json:"total_distance" bson:"total_distance"`
    }
)

type (  
    // User represents the structure of our resource
    UserResponse struct {
        Id bson.ObjectId `json:"id" bson:"_id"`
        Name string `json:"name" bson:"name"`
        Address string `json:"address" bson:"address"`
        City string `json:"city" bson:"city"`
        State string `json:"state" bson:"state"`
        Zip string `json:"zip" bson:"zip"`
        Coordinates AddressCoordinates
    }
)

type (  
    // User represents the structure of our resource
    AddressCoordinates struct {
        Lat float64 `json:"lat" bson:"lat"`
        Lng float64 `json:"lng" bson:"lng"`
    }
)

type (
    LocationAttributes struct {
        StartLatitude float64
        EndLatitude float64
        StartLongitude float64
        EndLongitude float64
    }    
)

// Uber price estimate
type (
    PriceEstimates struct {
        Prices []PriceEstimate `json:"prices"`
    }
)

// Uber price estimate
type (
    PriceEstimate struct {
        ProductId       string  `json:"product_id"`
        CurrencyCode    string  `json:"currency_code"`
        DisplayName     string  `json:"display_name"`
        Estimate        string  `json:"estimate"`
        LowEstimate     int     `json:"low_estimate"`
        HighEstimate    int     `json:"high_estimate"`
        SurgeMultiplier float64 `json:"surge_multiplier"`
        Duration        int     `json:"duration"`
        Distance        float64 `json:"distance"`
    }
)

// Uber put response
type (
    UberPutResponse struct {
        Status                          string `json:"status" bson:"status"`
        RequestId                       string `json:"request_id" bson:"request_id"`
        Driver DriverInformation        
        Eta                             int  `json:"eta" bson:"eta"`
        Location LocationInformation    
        Vehicle VehicleInformation      
        SurgeMultiplier                 float64 `json:"surge_multiplier"`
    }
)

type (
    DriverInformation struct {
        Phone           string `json:"phone_number" bson:"phone_number"`
        Rating          float64 `json:"rating" bson:"rating"`
        PictureUrl      string  `json:"picture_url" bson:"picture_url"`
        Name            string  `json:"name" bson:"name"`
    }
)

type (
    LocationInformation struct {
        Lat         float64 `json:"latitude" bson:"latitude"`
        Bearing     int `json:"bearing" bson:"bearing"`
        Lng         float64 `json:"longitude" bson:"longitude"`
    }
)

type (
    VehicleInformation struct {
        Make            string `json:"make" bson:"make"`
        PictureUrl      string `json:"picture_url" bson:"picture_url"`
        Model           string `json:"model" bson:"model"`
        LicensePlate    string `json:"license_plate" bson:"license_plate"`
    }
)


// Uber put response
type (
    UberStatusResponse struct {
        Id bson.ObjectId `json:"id" bson:"_id"`
        Status           string `json:"status" bson:"status"`
        StartingLocation string `json:"starting_from_location_id" bson:"starting_from_location_id"`
        NextDestination  string `json:"next_destination_location_id" bson:"next_destination_location_id"`
        BestRoute[] string `json:"best_route_location_ids" bson:"best_route_location_ids"`
        TotalCost int `json:"total_uber_costs" bson:"total_uber_costs"`
        TotalDuration int `json:"total_uber_duration" bson:"total_uber_duration"`
        TotalDistance float64 `json:"total_distance" bson:"total_distance"`
        Eta int `json:"uber_wait_time_eta" bson:"uber_wait_time_eta"`
    }
)
