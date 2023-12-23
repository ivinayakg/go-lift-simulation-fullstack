package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ivinayakg/go-lift-simulation/models"
	"github.com/ivinayakg/go-lift-simulation/services"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SessionCreateRequestBody struct {
	Floors int `json:"floors"`
	Lifts  int `json:"lifts"`
}

type LiftRequestCreateRequestBody struct {
	Floor int `json:"floor"`
}

var methodChoices = map[string]string{
	"get":   "GET",
	"post":  "POST",
	"patch": "PATCH",
	"del":   "DELETE",
}

func sendJSONError(w http.ResponseWriter, statusCode int, errorMessage string) {
	errorResponse := ErrorResponse{Error: errorMessage}
	w.Header().Set("content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

func setHeaders(type_ string, w http.ResponseWriter) {
	method := methodChoices[type_]
	if method == "" {
		method = "GET"
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if method != "GET" {
		w.Header().Set("Access-Control-Allow-Methods", method)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}
}

func CreateLift(w http.ResponseWriter, r *http.Request) {
	setHeaders("POST", w)
	var lift models.Lift
	json.NewDecoder(r.Body).Decode(&lift)
	if lift.Status == "" {
		lift.Status = "idle"
	}

	resultLift, err := models.CreateLift(lift)

	if err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(resultLift)
}

func GetLifts(w http.ResponseWriter, r *http.Request) {
	setHeaders("get", w)
	payload := models.GetLifts()
	json.NewEncoder(w).Encode(payload)
}

func CreateSession(w http.ResponseWriter, r *http.Request) {
	setHeaders("POST", w)
	var body SessionCreateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	floorsNumber := body.Floors
	liftsNumber := body.Lifts

	session, err := models.CreateSession(floorsNumber, liftsNumber)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	json.NewEncoder(w).Encode(session)
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	setHeaders("get", w)
	vars := mux.Vars(r)
	sessionID := vars["id"]
	payload, err := models.GetSession(sessionID)
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(payload)
}

func GetLiftRequests(w http.ResponseWriter, r *http.Request) {
	setHeaders("get", w)
	vars := mux.Vars(r)
	sessionID := vars["id"]
	payload, err := models.GetLiftRequests(sessionID, "")
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	json.NewEncoder(w).Encode(payload)
}

func CreateLiftRequest(w http.ResponseWriter, r *http.Request) {
	setHeaders("POST", w)
	var body LiftRequestCreateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	sessionID := vars["id"]

	floorNumber := body.Floor

	liftRequest, err := models.CreateLiftRequest(floorNumber, sessionID)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	services.Pubsubsys.AddToQue(liftRequest)
	json.NewEncoder(w).Encode(liftRequest)
}
