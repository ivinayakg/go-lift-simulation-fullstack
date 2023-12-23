package services

import (
	"log"

	"github.com/ivinayakg/go-lift-simulation/models"
)

type PubSub struct {
	Que        []*models.LiftRequest
	Processing bool
}

func NewPubSub() *PubSub {
	activeRequests, err := models.GetLiftRequests("", models.StatusQueued)

	if err != nil {
		log.Fatal(err)
	}

	return &PubSub{
		Que:        activeRequests,
		Processing: false,
	}
}

func (pubsub *PubSub) AddToQue(request *models.LiftRequest) {
	pubsub.Que = append(pubsub.Que, request)
}

func (pubsub *PubSub) EmptyQue(request *models.LiftRequest) {
	pubsub.Que = pubsub.Que[:0]
}

func (pubsub *PubSub) PopQue() *models.LiftRequest {
	first := pubsub.Que[0]
	pubsub.Que = pubsub.Que[1:]
	return first
}

func (pubsub *PubSub) Empty() bool {
	queLen := len(pubsub.Que)
	return queLen == 0
}

func (pubsub *PubSub) processRequests(cb func(*models.LiftRequest)) {
	if pubsub.Processing {
		return
	}

	pubsub.Processing = true

	for {
		queLen := pubsub.Empty()
		if !queLen {
			pubsub.Processing = false
			break
		}
		request := pubsub.PopQue()
		cb(request)
	}
}

var Pubsubsys = NewPubSub()
