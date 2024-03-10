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
	ENV := os.Getenv("ENV")

	if ENV != "production" || ENV == "" {
		envFileName := ".env.local"
		if ENV != "" && ENV != "development" {
			envFileName = ".env"
		}
		err := godotenv.Load(envFileName)
		if err != nil {
			log.Fatal("Error loading the .env file")
		}
	}

	router := mux.NewRouter()

	PORT := os.Getenv("PORT")
	allowed_origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), " ")
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: allowed_origins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})
	fmt.Println(allowed_origins)

	models.CreateDBInstance()
	services.SetupPubSub()

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
