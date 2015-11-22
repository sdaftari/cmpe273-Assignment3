package main 

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
    "net/http"
    "gopkg.in/mgo.v2"
    controller "Assignment3/controllers"    
)

func main() {
	fmt.Println("Server is listening on 8080!")

	connectionUserDb := controller.NewConnection(getSession())

    mux := httprouter.New()

    //Create new trip plan
    mux.POST("/trips", connectionUserDb.CreateTripPlan)

    //Get the trip plan
    mux.GET("/trips/:tripId", connectionUserDb.GetTripPlan)

    //Update the trip plan
    mux.PUT("/trips/:tripId/request", connectionUserDb.UpdateTripPlan)

    server := http.Server{
            Addr:        "0.0.0.0:8080",
            Handler: mux,
    }
    server.ListenAndServe()
}

func getSession() *mgo.Session {  
    // Connect to our local mongo
    s, err := mgo.Dial("mongodb://test:asdf1234#@ds041154.mongolab.com:41154/usersdb")

    // Check if connection error, is mongo running?
    if err != nil {
        panic(err)
    }
    return s
}