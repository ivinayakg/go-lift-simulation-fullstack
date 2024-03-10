package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/ivinayakg/go-lift-simulation/models"
	"github.com/ivinayakg/go-lift-simulation/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var SocketEvents = map[string]string{
	"User Joined": "user_joined",
	"Lift Moved":  "lift_moved",
	"User Left":   "user_left",
	"Client Info": "client_info",
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins (you might want to implement more specific logic)
		return true
	},
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, map[string]string, error) {
	queryParams := make(map[string]string)

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		return nil, queryParams, fmt.Errorf("session id is required")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, queryParams, err
	}
	queryParams["sessionID"] = sessionID
	return conn, queryParams, nil
}

type Client struct {
	ID          primitive.ObjectID
	Conn        *websocket.Conn
	Pool        *Pool
	SessionRoom *SessionRoom
}

type Message struct {
	Body      bson.M             `json:"body"`
	SessionID primitive.ObjectID `json:"session_id"`
	CreatedBy primitive.ObjectID `json:"created_by"`
}

type SessionRoom struct {
	SessionID primitive.ObjectID
	Clients   map[primitive.ObjectID]*Client
}

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Sessions   map[primitive.ObjectID]*SessionRoom
	Broadcast  chan *Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Sessions:   make(map[primitive.ObjectID]*SessionRoom),
		Broadcast:  make(chan *Message),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			clients := pool.Sessions[client.SessionRoom.SessionID].Clients
			client.Conn.WriteJSON(Message{Body: bson.M{"event": SocketEvents["Client Info"], "clientId": client.ID}, SessionID: client.SessionRoom.SessionID})
			fmt.Printf("\nSize of Connection Pool: %d, for the session ID %v\n", len(clients), client.SessionRoom.SessionID)
			for _, client := range clients {
				client.Conn.WriteJSON(Message{Body: bson.M{"event": SocketEvents["User Joined"]}, SessionID: client.SessionRoom.SessionID})
			}
		case client := <-pool.Unregister:
			clients := pool.Sessions[client.SessionRoom.SessionID].Clients
			for _, client := range clients {
				client.Conn.WriteJSON(Message{Body: bson.M{"event": SocketEvents["User Left"]}, SessionID: client.SessionRoom.SessionID})
			}
			delete(clients, client.ID)
			if len(clients) == 0 {
				delete(pool.Sessions, client.SessionRoom.SessionID)
			}
			fmt.Printf("\nSize of Connection Pool: %d, for the session ID %v\n", len(clients), client.SessionRoom.SessionID)
		case message := <-pool.Broadcast:
			if session := pool.Sessions[message.SessionID]; session != nil {
				clients := session.Clients
				fmt.Printf("\nSending message to all clients in session %v\n, message is %v", message.SessionID, message)
				for _, client := range clients {
					if client.ID == message.CreatedBy {
						continue
					}
					if err := client.Conn.WriteJSON(message); err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		}
	}
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func serveWS(pool *Pool, w http.ResponseWriter, r *http.Request) error {
	fmt.Println("WebSocket endpoint reached")

	conn, queryParams, err := Upgrade(w, r)
	if err != nil {
		return err
	}

	session, err := models.GetSession(queryParams["sessionID"])
	if err != nil {
		return err
	}

	sessionRoom := pool.Sessions[session.ID]
	if sessionRoom == nil {
		sessionRoom = &SessionRoom{
			SessionID: session.ID,
			Clients:   make(map[primitive.ObjectID]*Client),
		}
		pool.Sessions[session.ID] = sessionRoom
	}

	clientUUID := utils.GenerateUUID()
	client := &Client{
		Conn:        conn,
		Pool:        pool,
		ID:          clientUUID,
		SessionRoom: sessionRoom,
	}

	sessionRoom.Clients[clientUUID] = client
	pool.Register <- client
	client.Read()

	return nil
}

func DeployWS(router *mux.Router) *Pool {
	pool := NewPool()
	go pool.Start()

	router.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		if err := serveWS(pool, w, r); err != nil {
			fmt.Fprintf(w, "%+v\n", err)
		}
	})

	return pool
}
