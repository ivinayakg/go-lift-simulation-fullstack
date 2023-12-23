package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/ivinayakg/go-lift-simulation/controllers"
	"github.com/rs/cors"
)

const PORT = "5454"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("websocket ban gaya")
	reader(ws)
}

func main() {
	router := mux.NewRouter()

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:19006"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	})

	// router.HandleFunc("/lift", controllers.CreateLift).Methods("POST", "OPTIONS")
	router.HandleFunc("/session", controllers.CreateSession).Methods("POST", "OPTIONS")
	router.HandleFunc("/session/{id}", controllers.GetSession).Methods("GET", "OPTIONS")
	router.HandleFunc("/session/{id}/request", controllers.CreateLiftRequest).Methods("POST", "OPTIONS")
	router.HandleFunc("/session/{id}/request", controllers.GetLiftRequests).Methods("GET", "OPTIONS")
	router.HandleFunc("/ws", wsEndpoint)
	// router.HandleFunc("/lift", controllers.GetLifts).Methods("GET", "OPTIONS")

	routerProtected := corsHandler.Handler(router)

	fmt.Println("Starting the server on port " + PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", PORT), routerProtected))
}
