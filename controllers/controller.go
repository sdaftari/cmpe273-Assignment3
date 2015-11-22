package Controller

import (
	"encoding/json"
    "fmt"
    "io"
    "strconv"
    "io/ioutil"
    "net/http"
	"github.com/julienschmidt/httprouter"
    "bytes"

	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"

	model "Assignment3/models"
)

var StartingLocationFinal string 
var BestRouteFinal []string
var TotalCostFinal int
var TotalDurationFinal int
var TotalDistanceFinal float64
var isFirstDestination bool
var isLastDestinationSource bool
var isLastRequest bool

const (
    // Uber API endpoint
    ServerToken string = "wLxxfO8eKBD_fi1C3q6P0DM3XOKIDMvMTR2hLZXN"
    UberLocation string = "https://api.uber.com/v1/estimates/price?"
    ProductIdConst string = "d4abaae7-f4d6-4152-91cc-77523e8165a4"
)

type (  
    // UserController represents the controller for operating on the User resource
    ConnectionUserDb struct{
        session *mgo.Session
    }
)

// Uber price estimate
type (
    DistancePriceArray struct {
        sortedDistanceArray []DistancePriceObject 
    }
)

// Uber price estimate
type (
    DistancePriceObject struct {
        locationId       string  
        price            int
        duration         int  
    }
)

func NewConnection(s *mgo.Session) *ConnectionUserDb {  
    //retur the object of UserController
    return &ConnectionUserDb{s}
}

func (uc ConnectionUserDb) GetTripPlan(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    tripResponse := model.TripResponse{}

    id := p.ByName("tripId")

    if !bson.IsObjectIdHex(id) {
        rw.WriteHeader(404)
        return
    }

    objectId := bson.ObjectIdHex(id)

    // Fetch user
    if err := uc.session.DB("usersdb").C("trips").FindId(objectId).One(&tripResponse); err != nil {
        rw.WriteHeader(404)
        return
    }

    // Marshal provided interface into JSON structure
    uj, _ := json.Marshal(tripResponse)

    // Write content-type, statuscode, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", uj)
}

func (uc ConnectionUserDb) UpdateTripPlan(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    tripResponse := model.TripResponse{}
    userResponse := model.UserResponse{}
    uberPutResponse := model.UberPutResponse{}
    uberStatusResponse := model.UberStatusResponse{}
    var nextDestination string
    var startLocationId string

    //Get the locationId
    tripId := p.ByName("tripId")

    if !bson.IsObjectIdHex(tripId) {
        rw.WriteHeader(404)
        return
    }

    objectId := bson.ObjectIdHex(tripId)

    //Get the plan
    if err := uc.session.DB("usersdb").C("trips").FindId(objectId).One(&tripResponse); err != nil {
        rw.WriteHeader(404)
        return
    }

    if isLastRequest {
        uberStatusResponse.Id = tripResponse.Id  
        uberStatusResponse.StartingLocation = StartingLocationFinal
        uberStatusResponse.NextDestination = ""
        uberStatusResponse.Status = "finished"
        uberStatusResponse.BestRoute = BestRouteFinal
        uberStatusResponse.TotalCost = TotalCostFinal
        uberStatusResponse.TotalDuration = TotalDurationFinal
        uberStatusResponse.TotalDistance = TotalDistanceFinal
        uberStatusResponse.Eta = 0

        uj, _ := json.Marshal(uberStatusResponse)

        // Write content-type, statuscode, payload
        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(200)
        fmt.Fprintf(rw, "%s", uj)
    } else {
        if isFirstDestination {
            startLocationId = StartingLocationFinal
        } else {
            startLocationId = tripResponse.BestRoute[0]
        } 

        if !bson.IsObjectIdHex(startLocationId) {
            rw.WriteHeader(404)
            return
        }

        objectIdLocation := bson.ObjectIdHex(startLocationId)

        //Get the starting location information
        if err := uc.session.DB("usersdb").C("userAddresses").FindId(objectIdLocation).One(&userResponse); err != nil {
            rw.WriteHeader(404)
            return
        }

        startLatitude := userResponse.Coordinates.Lat
        startLongitude := userResponse.Coordinates.Lng

        if len(tripResponse.BestRoute) == 1 {
            isLastDestinationSource = true
            isLastRequest = true
        }
        if isFirstDestination {
            nextDestination = tripResponse.BestRoute[0]
        } else {
            if isLastDestinationSource {
                nextDestination = StartingLocationFinal
            } else {
                nextDestination = tripResponse.BestRoute[1]
            }            
        }        

        if !bson.IsObjectIdHex(nextDestination) {
            rw.WriteHeader(404)
            return
        }

        objectIdLocation = bson.ObjectIdHex(nextDestination)

        //Get the starting location information
        if err := uc.session.DB("usersdb").C("userAddresses").FindId(objectIdLocation).One(&userResponse); err != nil {
            rw.WriteHeader(404)
            return
        }

        destLatitude := userResponse.Coordinates.Lat
        destLongitude := userResponse.Coordinates.Lng

        urlPost := "https://sandbox-api.uber.com/v1/requests?start_latitude=" + strconv.FormatFloat(startLatitude, 'f', 6, 64) + "&start_longitude=" + strconv.FormatFloat(startLongitude, 'f', 6, 64) + "&end_latitude=" + strconv.FormatFloat(destLatitude, 'f', 6, 64) + "&end_longitude=" + strconv.FormatFloat(destLongitude, 'f', 6, 64) + "&product_id=" + ProductIdConst
        requestBody := []byte(`{"start_longitude":"` + strconv.FormatFloat(userResponse.Coordinates.Lng, 'f', 6, 64) + `", "start_latitude":"` + strconv.FormatFloat(userResponse.Coordinates.Lat, 'f', 6, 64) + `", "product_id":"` + ProductIdConst + `"}`)
        
        reqUber, err := http.NewRequest("POST", urlPost, bytes.NewBuffer(requestBody))
        reqUber.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsInJlcXVlc3RfcmVjZWlwdCIsInJlcXVlc3QiLCJoaXN0b3J5X2xpdGUiXSwic3ViIjoiM2JiOTgyZGMtNjE1Zi00YjFiLWEwOGYtZTMwZDhlOTE3YjYzIiwiaXNzIjoidWJlci11czEiLCJqdGkiOiJhYTk2YThmZi1jM2I2LTRlNzktOWJjNi05NzFkYzA3ZDk0MGEiLCJleHAiOjE0NTA1NTM4MjksImlhdCI6MTQ0Nzk2MTgyOCwidWFjdCI6IkR4U21oQWFNaW1tRDhjWFlpVEJiVndZdXVScW81SSIsIm5iZiI6MTQ0Nzk2MTczOCwiYXVkIjoiTUIxVllOMV9Jb0d0RTUzMmZBcVcxejhDNTN2YklteWMifQ.GNvzC8tLo4UhzpJr_Bkg0c7XcA2XJhaiQn58LNQcdRFr5eD9hXLos1AFkC2u-IsYjV64LCzShaopyNN-NwezfabbzHS0qjXqyXBFfWbKSlsYLpNpF6ethrRQAq3PhwQGDXqNpEFkBAleQKMQj_lK5Mxo5IfG6F5Fm6b-GsWM6RQcWl8KcnLi_ZejAZrkC78lXL_aIfPSAXDsubhkJs-dR-T6ptj_GFl8DyTlA14PUnHZ_TTK9ic-yOZUBa1rNEHutOlDRGmC-UqHt01e5c4Lau7CGp80gLetn4UxX27s3azhu1zf0MwHTs560T4nli9PyM9dwdgbqLr4IIKr-6c10A")
        reqUber.Header.Set("Content-Type", "application/json")

        client := &http.Client{}
        resp, err := client.Do(reqUber)
        if err != nil {
            panic(err)
        }
        defer resp.Body.Close()

        bodyResp, err1 := ioutil.ReadAll(resp.Body)
        if err1 != nil {
            panic(err)
        }
        
        err = json.Unmarshal(bodyResp, &uberPutResponse)
        if err != nil {
            panic(err)
        }

        requestBody = []byte(`{"status":"accepted"}`)
        urlPut := "https://sandbox-api.uber.com/v1/sandbox/requests/" + uberPutResponse.RequestId
        reqUber, err = http.NewRequest("PUT", urlPut, bytes.NewBuffer(requestBody))
        reqUber.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsInJlcXVlc3RfcmVjZWlwdCIsInJlcXVlc3QiLCJoaXN0b3J5X2xpdGUiXSwic3ViIjoiM2JiOTgyZGMtNjE1Zi00YjFiLWEwOGYtZTMwZDhlOTE3YjYzIiwiaXNzIjoidWJlci11czEiLCJqdGkiOiJhYTk2YThmZi1jM2I2LTRlNzktOWJjNi05NzFkYzA3ZDk0MGEiLCJleHAiOjE0NTA1NTM4MjksImlhdCI6MTQ0Nzk2MTgyOCwidWFjdCI6IkR4U21oQWFNaW1tRDhjWFlpVEJiVndZdXVScW81SSIsIm5iZiI6MTQ0Nzk2MTczOCwiYXVkIjoiTUIxVllOMV9Jb0d0RTUzMmZBcVcxejhDNTN2YklteWMifQ.GNvzC8tLo4UhzpJr_Bkg0c7XcA2XJhaiQn58LNQcdRFr5eD9hXLos1AFkC2u-IsYjV64LCzShaopyNN-NwezfabbzHS0qjXqyXBFfWbKSlsYLpNpF6ethrRQAq3PhwQGDXqNpEFkBAleQKMQj_lK5Mxo5IfG6F5Fm6b-GsWM6RQcWl8KcnLi_ZejAZrkC78lXL_aIfPSAXDsubhkJs-dR-T6ptj_GFl8DyTlA14PUnHZ_TTK9ic-yOZUBa1rNEHutOlDRGmC-UqHt01e5c4Lau7CGp80gLetn4UxX27s3azhu1zf0MwHTs560T4nli9PyM9dwdgbqLr4IIKr-6c10A")
        reqUber.Header.Set("Content-Type", "application/json")

        //client = &http.Client{}
        resp, err = client.Do(reqUber)
        if err != nil {
            panic(err)
        }
        defer resp.Body.Close()

        bodyResp, err1 = ioutil.ReadAll(resp.Body)
        if err1 != nil {
            panic(err)
        }

        if isLastDestinationSource {
            requestBody = []byte(`{"start_longitude":"` + strconv.FormatFloat(userResponse.Coordinates.Lng, 'f', 6, 64) + `", "start_latitude":"` + strconv.FormatFloat(userResponse.Coordinates.Lat, 'f', 6, 64) + `", "product_id":"` + "a1111c8c-c720-46c3-8534-2fcdd730040d" + `"}`) 
        } else {
            requestBody = []byte(`{"start_longitude":"` + strconv.FormatFloat(userResponse.Coordinates.Lng, 'f', 6, 64) + `", "start_latitude":"` + strconv.FormatFloat(userResponse.Coordinates.Lat, 'f', 6, 64) + `", "product_id":"` + ProductIdConst + `"}`)
        }
        
        reqUber, err = http.NewRequest("POST", urlPost, bytes.NewBuffer(requestBody))
        reqUber.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsInJlcXVlc3RfcmVjZWlwdCIsInJlcXVlc3QiLCJoaXN0b3J5X2xpdGUiXSwic3ViIjoiM2JiOTgyZGMtNjE1Zi00YjFiLWEwOGYtZTMwZDhlOTE3YjYzIiwiaXNzIjoidWJlci11czEiLCJqdGkiOiJhYTk2YThmZi1jM2I2LTRlNzktOWJjNi05NzFkYzA3ZDk0MGEiLCJleHAiOjE0NTA1NTM4MjksImlhdCI6MTQ0Nzk2MTgyOCwidWFjdCI6IkR4U21oQWFNaW1tRDhjWFlpVEJiVndZdXVScW81SSIsIm5iZiI6MTQ0Nzk2MTczOCwiYXVkIjoiTUIxVllOMV9Jb0d0RTUzMmZBcVcxejhDNTN2YklteWMifQ.GNvzC8tLo4UhzpJr_Bkg0c7XcA2XJhaiQn58LNQcdRFr5eD9hXLos1AFkC2u-IsYjV64LCzShaopyNN-NwezfabbzHS0qjXqyXBFfWbKSlsYLpNpF6ethrRQAq3PhwQGDXqNpEFkBAleQKMQj_lK5Mxo5IfG6F5Fm6b-GsWM6RQcWl8KcnLi_ZejAZrkC78lXL_aIfPSAXDsubhkJs-dR-T6ptj_GFl8DyTlA14PUnHZ_TTK9ic-yOZUBa1rNEHutOlDRGmC-UqHt01e5c4Lau7CGp80gLetn4UxX27s3azhu1zf0MwHTs560T4nli9PyM9dwdgbqLr4IIKr-6c10A")
        reqUber.Header.Set("Content-Type", "application/json")

        resp, err = client.Do(reqUber)
        if err != nil {
            panic(err)
        }
        defer resp.Body.Close()

        bodyResp, err1 = ioutil.ReadAll(resp.Body)
        if err1 != nil {
            panic(err)
        }

        err = json.Unmarshal(bodyResp, &uberPutResponse)
        if err != nil {
            panic(err)
        }

        uberStatusResponse.Id = tripResponse.Id 
        uberStatusResponse.StartingLocation = tripResponse.StartingLocation
        uberStatusResponse.NextDestination = nextDestination
        uberStatusResponse.Status = "requesting"
        uberStatusResponse.BestRoute = BestRouteFinal
        uberStatusResponse.TotalCost = tripResponse.TotalCost
        uberStatusResponse.TotalDuration = tripResponse.TotalDuration
        uberStatusResponse.TotalDistance = tripResponse.TotalDistance
        uberStatusResponse.Eta = uberPutResponse.Eta

        uj, _ := json.Marshal(uberStatusResponse)

        // Write content-type, statuscode, payload
        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(200)
        fmt.Fprintf(rw, "%s", uj)

        tripResponse.BestRoute = append(tripResponse.BestRoute[:0], tripResponse.BestRoute[1:]...)   
                
        if !isFirstDestination {
            err1 = uc.session.DB("usersdb").C("trips").Update(bson.M{"_id":objectId }, 
                bson.M{"$set": bson.M{"status": "requesting", "starting_from_location_id": StartingLocationFinal, "best_route_location_ids": tripResponse.BestRoute, "total_uber_costs": tripResponse.TotalCost, "total_uber_duration": tripResponse.TotalDuration, "total_uber_distance": tripResponse.TotalDistance, "uber_wait_time_eta": uberPutResponse.Eta}})
        }

        isFirstDestination = false 
        isLastDestinationSource = false 
    }            
}

func (uc ConnectionUserDb) CreateTripPlan(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	tripRequest := model.TripRequest{}
    tripResponse := model.TripResponse{}
    uberEstimate := model.PriceEstimates{}
    userResponse := model.UserResponse{}
    locationAttributes := model.LocationAttributes{}   
    objDistancePriceArray := DistancePriceArray{} 

    isFirstDestination = true
    isLastDestinationSource = false
    isLastRequest = false

    var totalCost int
    var totalDuration int
    var totalDistance float64

    //Get data from request
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
    }

    err = json.Unmarshal(body, &tripRequest)
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
    }

    //Get the locationId
    id := tripRequest.StartingLocation
    StartingLocationFinal = id

    if !bson.IsObjectIdHex(id) {
        rw.WriteHeader(404)
        return
    }

    objectId := bson.ObjectIdHex(id)

    //Get the starting location information
    if err := uc.session.DB("usersdb").C("userAddresses").FindId(objectId).One(&userResponse); err != nil {
        rw.WriteHeader(404)
        return
    }    

    //Set the starting lattitude and longitude
    locationAttributes.StartLatitude = userResponse.Coordinates.Lat
    locationAttributes.StartLongitude = userResponse.Coordinates.Lng

    ids := tripRequest.LocationIds
    startLat := locationAttributes.StartLatitude
    startLon := locationAttributes.StartLongitude
    var destination []string
    var locationIdsArray []string
    var tempIds []string

    count := len(ids)
    
    //Get the best route based on shortest path and cost
    for i := 0; i < count; i++ {
        objDistancePriceArray.sortedDistanceArray = GetNearestDestinationId(startLat, startLon, ids)
        destination = append(destination, objDistancePriceArray.sortedDistanceArray[0].locationId)
        tempIds = nil
        for j := 0; j < len(objDistancePriceArray.sortedDistanceArray); j++ {
            tempIds = append(tempIds, objDistancePriceArray.sortedDistanceArray[j].locationId)
        }

        ids = tempIds

        //Retrieve information about each destination id
        session, err := mgo.Dial("mongodb://test:asdf1234#@ds041154.mongolab.com:41154/usersdb")
        if err != nil {
            panic(err)
        }
        defer session.Close()

        // Optional. Switch the session to a monotonic behavior.
        session.SetMode(mgo.Monotonic, true)

        objectId := bson.ObjectIdHex(ids[0])

        err1 := session.DB("usersdb").C("userAddresses").FindId(objectId).One(&userResponse)
        if(err1 != nil) {
            fmt.Println(err1)
        }

        //Set lattitude and longitude of destination
        startLat = userResponse.Coordinates.Lat
        startLon = userResponse.Coordinates.Lng

        if(i < count) {
            ids = append(ids[:0], ids[1:]...)
        }        
    }

    tripResponse.Id = bson.NewObjectId()
    tripResponse.Status = "planning"
    tripResponse.StartingLocation = id
    tripResponse.BestRoute = destination

    BestRouteFinal = destination

    //Append source id for round trip
    locationIdsArray = append(locationIdsArray, tripRequest.StartingLocation)
    for i := 0; i < len(destination); i++ {
        locationIdsArray = append(locationIdsArray, destination[i])
    }
    locationIdsArray = append(locationIdsArray, tripRequest.StartingLocation)

    for i := 0; i < len(locationIdsArray)-1; i++ {
        startlocationId := locationIdsArray[i]

        if !bson.IsObjectIdHex(startlocationId) {
            rw.WriteHeader(404)
            return
        }

        objectIdLocation := bson.ObjectIdHex(startlocationId)

        //Retrieve information about each destination id
        if err := uc.session.DB("usersdb").C("userAddresses").FindId(objectIdLocation).One(&userResponse); err != nil {
            rw.WriteHeader(404)
            return
        }

        //Set Start Location
        locationAttributes.StartLatitude = userResponse.Coordinates.Lat
        locationAttributes.StartLongitude = userResponse.Coordinates.Lng

        endLocationId := locationIdsArray[i + 1]

        if !bson.IsObjectIdHex(endLocationId) {
            rw.WriteHeader(404)
            return
        }

        objectIdLocation = bson.ObjectIdHex(endLocationId)

        //Retrieve information about each destination id
        if err := uc.session.DB("usersdb").C("userAddresses").FindId(objectIdLocation).One(&userResponse); err != nil {
            rw.WriteHeader(404)
            return
        }
        
        //Set lattitude and longitude of destination
        locationAttributes.EndLatitude = userResponse.Coordinates.Lat
        locationAttributes.EndLongitude = userResponse.Coordinates.Lng  

        uberResponse, err := GetUberPriceEstimation(locationAttributes)
        uberResponseBody, err := ioutil.ReadAll(uberResponse)
        err = json.Unmarshal(uberResponseBody, &uberEstimate)

        if err != nil {
            fmt.Println(err.Error())
            return
        }

        totalCost = totalCost + uberEstimate.Prices[0].LowEstimate
        totalDuration = totalDuration + uberEstimate.Prices[0].Duration
        totalDistance = totalDistance + uberEstimate.Prices[0].Distance
    }

    tripResponse.TotalCost = totalCost
    tripResponse.TotalDuration = totalDuration
    tripResponse.TotalDistance = totalDistance

    uj, _ := json.Marshal(tripResponse)

    // Write content-type, statuscode, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", uj)

    TotalCostFinal = totalCost
    TotalDurationFinal = totalDuration
    TotalDistanceFinal = totalDistance

    //Save the trip plan in database
    uc.session.DB("usersdb").C("trips").Insert(tripResponse)
}

func GetNearestDestinationId(startingLat float64, startingLon float64, ids []string) []DistancePriceObject {
    locationAttributes := model.LocationAttributes{}
    userResponse := model.UserResponse{}
    uberEstimate := model.PriceEstimates{}
    objDistancePrice := DistancePriceObject{}
    tempStruct := DistancePriceObject{}
    objDistancePriceArray := DistancePriceArray{}

    locationAttributes.StartLatitude = startingLat
    locationAttributes.StartLongitude = startingLon

    for _, i := range ids {
        locationId := i

        if !bson.IsObjectIdHex(locationId) {
            return nil
        }

        objectId := bson.ObjectIdHex(locationId)

        //Retrieve information about each destination id
        session, err := mgo.Dial("mongodb://test:asdf1234#@ds041154.mongolab.com:41154/usersdb")
        if err != nil {
            panic(err)
        }
        defer session.Close()

        // Optional. Switch the session to a monotonic behavior.
        session.SetMode(mgo.Monotonic, true)

        err1 := session.DB("usersdb").C("userAddresses").FindId(objectId).One(&userResponse)
        if(err1 != nil) {
            panic(err)
        }
        defer session.Close()

        //Set lattitude and longitude of destination
        locationAttributes.EndLatitude = userResponse.Coordinates.Lat
        locationAttributes.EndLongitude = userResponse.Coordinates.Lng  

        uberResponse, err2 := GetUberPriceEstimation(locationAttributes)
        if(err2 != nil) {
            panic(err)
        }
        defer session.Close()
        uberResponseBody, err3 := ioutil.ReadAll(uberResponse)
        if(err3 != nil) {
            panic(err)
        }
        defer session.Close()

        err = json.Unmarshal(uberResponseBody, &uberEstimate)
        if err != nil {
            fmt.Println(err.Error())
            return nil
        }  
        objDistancePrice.locationId = locationId
        objDistancePrice.price = uberEstimate.Prices[0].LowEstimate
        objDistancePrice.duration = uberEstimate.Prices[0].Duration

        objDistancePriceArray.AddItem(objDistancePrice)
    } 

    for i := 0; i < len(objDistancePriceArray.sortedDistanceArray); i++ {
        for j := i+1; j < len(objDistancePriceArray.sortedDistanceArray); j++ {
            if objDistancePriceArray.sortedDistanceArray[i].price > objDistancePriceArray.sortedDistanceArray[j].price {
                    tempStruct = objDistancePriceArray.sortedDistanceArray[i]
                    objDistancePriceArray.sortedDistanceArray[i] = objDistancePriceArray.sortedDistanceArray[j]
                    objDistancePriceArray.sortedDistanceArray[j] = tempStruct
                } else if objDistancePriceArray.sortedDistanceArray[i].price == objDistancePriceArray.sortedDistanceArray[j].price {
                    if objDistancePriceArray.sortedDistanceArray[i].duration > objDistancePriceArray.sortedDistanceArray[j].duration {
                        tempStruct = objDistancePriceArray.sortedDistanceArray[i]
                        objDistancePriceArray.sortedDistanceArray[i] = objDistancePriceArray.sortedDistanceArray[j]
                        objDistancePriceArray.sortedDistanceArray[j] = tempStruct
                    }
                }
        }
    }

    return objDistancePriceArray.sortedDistanceArray
}

func (distancePriceArray *DistancePriceArray) AddItem(item DistancePriceObject) []DistancePriceObject {
    distancePriceArray.sortedDistanceArray = append(distancePriceArray.sortedDistanceArray, item)
    return distancePriceArray.sortedDistanceArray
}

func GetUberPriceEstimation(locationAttributes model.LocationAttributes) (io.ReadCloser, error) {
    location := UberLocation + "start_latitude=" + strconv.FormatFloat(locationAttributes.StartLatitude, 'f', 6, 64) + 
                "&start_longitude=" + strconv.FormatFloat(locationAttributes.StartLongitude, 'f', 6, 64) + 
                "&end_latitude=" + strconv.FormatFloat(locationAttributes.EndLatitude, 'f', 6, 64) + 
                "&end_longitude=" + strconv.FormatFloat(locationAttributes.EndLongitude, 'f', 6, 64) + 
                "&server_token=" + ServerToken

    //Get response from uber apiLocaton:  
    response, err := http.Get(location)
    if err != nil {
        return nil, err
    }

    return response.Body, nil;
}

