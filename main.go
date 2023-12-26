package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ivinayakg/go-lift-simulation/controllers"
	"github.com/ivinayakg/go-lift-simulation/models"
	"github.com/ivinayakg/go-lift-simulation/services"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading the ENV file")
	}

	PORT := os.Getenv("PORT")
	allowed_origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), " ")
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: allowed_origins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	router := mux.NewRouter()

	router.HandleFunc("/session", controllers.CreateSession).Methods("POST", "OPTIONS")
	router.HandleFunc("/session/{id}", controllers.GetSession).Methods("GET", "OPTIONS")
	router.HandleFunc("/session/{id}/request", controllers.CreateLiftRequest).Methods("POST", "OPTIONS")
	router.HandleFunc("/session/{id}/request/", controllers.GetLiftRequests).Methods("GET", "OPTIONS")
	pool := services.DeployWS(router)

	routerProtected := corsHandler.Handler(router)

	go services.Pubsubsys.ProcessRequests(func(lr *services.LiftRequestEvent) {
		pool.Broadcast <- &services.Message{SessionID: lr.Session, Body: bson.M{"event": services.SocketEvents["Lift Moved"], "floor_requested": lr.RequestedFloor, "lift_id": lr.Lift}, CreatedBy: lr.CreatedBy}
		go func() {
			time.Sleep(3 * time.Second)
			requestObject := &models.LiftRequest{ID: lr.ID, RequestedFloor: lr.RequestedFloor, Lift: lr.Lift, Status: lr.Status, Session: lr.Session}
			models.CompleteLiftRequest(requestObject)
		}()
	})

	fmt.Println("Starting the server on port " + PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", PORT), routerProtected))
}
