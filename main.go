package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ivinayakg/go-lift-simulation/controllers"
	"github.com/ivinayakg/go-lift-simulation/models"
	"github.com/ivinayakg/go-lift-simulation/services"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
)

const PORT = "5454"

func main() {
	router := mux.NewRouter()

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:19006"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	})

	router.HandleFunc("/session", controllers.CreateSession).Methods("POST", "OPTIONS")
	router.HandleFunc("/session/{id}", controllers.GetSession).Methods("GET", "OPTIONS")
	router.HandleFunc("/session/{id}/request", controllers.CreateLiftRequest).Methods("POST", "OPTIONS")
	router.HandleFunc("/session/{id}/request/", controllers.GetLiftRequests).Methods("GET", "OPTIONS")
	pool := services.DeployWS(router)

	routerProtected := corsHandler.Handler(router)

	go services.Pubsubsys.ProcessRequests(func(lr *services.LiftRequestEvent) {
		pool.Broadcast <- &services.Message{SessionID: lr.Session, Body: bson.M{"event": services.SocketEvents["Lift Moved"], "floor_requested": lr.RequestedFloor, "lift_id": lr.Lift}, CreatedBy: lr.CreatedBy}
		go func() {
			time.Sleep(10 * time.Second)
			requestObject := &models.LiftRequest{ID: lr.ID, RequestedFloor: lr.RequestedFloor, Lift: lr.Lift, Status: lr.Status, Session: lr.Session}
			models.CompleteLiftRequest(requestObject)
		}()
	})

	fmt.Println("Starting the server on port " + PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", PORT), routerProtected))
}
