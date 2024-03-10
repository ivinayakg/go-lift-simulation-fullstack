package services

import (
	"fmt"
	"log"

	"github.com/ivinayakg/go-lift-simulation/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LiftRequestEvent struct {
	ID             primitive.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	RequestedFloor int                `json:"requestedFloor"`
	Lift           primitive.ObjectID `json:"lift"`
	Status         string             `json:"status,omitempty"`
	Session        primitive.ObjectID `json:"session"`
	CreatedBy      primitive.ObjectID `json:"created_by"`
}

type PubSub struct {
	Que         chan *LiftRequestEvent
	Processing  bool
	QueLength   int
	QueCapacity int
}

func NewPubSub() *PubSub {
	activeRequests, err := models.GetLiftRequests("", models.StatusQueued)
	if err != nil {
		log.Fatal(err)
	}

	queCapacity := 108
	queChannel := make(chan *LiftRequestEvent, queCapacity)
	queLength := len(activeRequests)

	for _, request := range activeRequests {
		fmt.Println(request)
		request := &LiftRequestEvent{ID: request.ID, Lift: request.Lift, RequestedFloor: request.RequestedFloor, Status: request.Status, Session: request.Session, CreatedBy: primitive.NilObjectID}
		queChannel <- request
	}

	return &PubSub{
		Que:         queChannel,
		Processing:  false,
		QueLength:   queLength,
		QueCapacity: queCapacity,
	}
}

func (pubsub *PubSub) AddToQue(request *LiftRequestEvent) {
	pubsub.Que <- request
	pubsub.QueLength++
}

func (pubsub *PubSub) PopQue() *LiftRequestEvent {
	if pubsub.QueLength > 0 {
		request := <-pubsub.Que
		pubsub.QueLength--
		return request
	}
	return nil
}

func (pubsub *PubSub) Empty() bool {
	return pubsub.QueLength == 0
}

func (pubsub *PubSub) ProcessRequests(cb func(*LiftRequestEvent)) {
	// if pubsub.Processing {
	// 	return
	// }

	// pubsub.Processing = true

	for {
		select {
		case request := <-pubsub.Que:
			// queLen := pubsub.Empty()
			// if !queLen {
			// 	pubsub.Processing = false
			// 	break
			// }
			pubsub.QueLength--
			cb(request)
		}
	}
}

var Pubsubsys *PubSub

func SetupPubSub() {
	Pubsubsys = NewPubSub()
}
