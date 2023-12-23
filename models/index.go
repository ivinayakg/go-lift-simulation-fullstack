package models

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ivinayakg/go-lift-simulation/utils"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// type User struct {
// 	ID     bson.ObjectId `json:"id" bson:"id"`
// 	Name   string        `json:"name" bson:"name"`
// 	Gender string        `json:"gender" bson:"gender"`
// 	Age    int           `json:"age" bson:"age"`
// }

type Lift struct {
	ID           primitive.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	CurrentFloor int                `json:"currentFloor"`
	Status       string             `json:"status,omitempty"`
}

type Session struct {
	ID     primitive.ObjectID   `json:"_id,omitempty"  bson:"_id,omitempty"`
	Lifts  []primitive.ObjectID `json:"lifts"`
	Floors int                  `json:"floors"`
}

type SessionResponse struct {
	ID     primitive.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	Lifts  []Lift             `json:"lifts"`
	Floors int                `json:"floors"`
}

type LiftRequest struct {
	ID             primitive.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	RequestedFloor int                `json:"requestedFloor"`
	Lift           primitive.ObjectID `json:"lift"`
	Status         string             `json:"status,omitempty"`
	Session        primitive.ObjectID `json:"session"`
}

const (
	StatusIdle      = "idle"
	StatusBusy      = "busy"
	StatusQueued    = "queued"
	StatusCompleted = "completed"
)

func ValidateLiftRequestStatus(status string) error {
	status = strings.ToLower(status)
	switch status {
	case StatusQueued, StatusCompleted:
		return nil // Status is valid.
	default:
		return errors.New("invalid status, valid status are queued, completed")
	}
}
func ValidateLiftStatus(status string) error {
	status = strings.ToLower(status)
	switch status {
	case StatusIdle, StatusBusy:
		return nil // Status is valid.
	default:
		return errors.New("invalid status, valid status are idle, busy")
	}
}

var liftCollection *mongo.Collection
var liftRequestCollection *mongo.Collection
var sessionCollection *mongo.Collection

func init() {
	loadTheEnv()
	createDBInstance()
}

func loadTheEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading the .env file")
	}
}

func createDBInstance() {
	connectionString := os.Getenv("DB_URI")
	dbName := os.Getenv("DB_NAME")
	liftCollName := os.Getenv("DB_LIFT_COLLECTION_NAME")
	liftRequestCollName := os.Getenv("DB_LIFT_REQUEST_COLLECTION_NAME")
	sessionCollName := os.Getenv("DB_SESSION_COLLECTION_NAME")

	clientOptions := options.Client().ApplyURI(connectionString)

	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to mongodb")

	liftCollection = client.Database(dbName).Collection(liftCollName)
	liftRequestCollection = client.Database(dbName).Collection(liftRequestCollName)
	sessionCollection = client.Database(dbName).Collection(sessionCollName)
}

func CreateLift(lift Lift) (primitive.ObjectID, error) {
	if err := ValidateLiftStatus(lift.Status); err != nil {
		return primitive.NilObjectID, err
	}

	result, err := liftCollection.InsertOne(context.Background(), lift)

	if err != nil {
		log.Fatal(err)
		return primitive.NilObjectID, err
	}

	fmt.Println("added a lift", result.InsertedID)
	return result.InsertedID.(primitive.ObjectID), nil
}

func GetLifts() []primitive.M {
	curr, err := liftCollection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	var results []primitive.M
	for curr.Next(context.Background()) {
		var result bson.M
		e := curr.Decode(&result)
		if e != nil {
			log.Fatal(e)
		}
		results = append(results, result)
	}

	if err := curr.Err(); err != nil {
		log.Fatal(err)
	}

	curr.Close(context.Background())
	return results
}

func CreateSession(floors int, lifts int) (*SessionResponse, error) {
	var liftObjs []interface{}
	for i := 0; i < lifts; i++ {
		lift := Lift{CurrentFloor: 0, Status: "idle"}
		liftObjs = append(liftObjs, lift)
	}

	liftResults, err := liftCollection.InsertMany(context.Background(), liftObjs)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Create a slice to store the inserted lift IDs.
	var insertedLiftIDs []primitive.ObjectID
	for _, result := range liftResults.InsertedIDs {
		insertedLiftIDs = append(insertedLiftIDs, result.(primitive.ObjectID))
	}

	// Create the Session object with the inserted lift IDs.
	session := Session{Floors: floors, Lifts: insertedLiftIDs}
	result, err := sessionCollection.InsertOne(context.Background(), session)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var insertedLifts []Lift
	for _, id := range insertedLiftIDs {
		insertedLifts = append(insertedLifts, Lift{ID: id, CurrentFloor: 0, Status: "idle"})
	}

	session.ID = result.InsertedID.(primitive.ObjectID)

	sessionResponse := SessionResponse{ID: session.ID, Floors: session.Floors, Lifts: insertedLifts}

	return &sessionResponse, nil
}

func GetSession(sessionID string) (*SessionResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		log.Fatal(err)
	}

	sessionFilter := bson.M{"_id": objectID}
	var session Session

	err = sessionCollection.FindOne(context.TODO(), sessionFilter).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Session Document not found")
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Found document: %+v\n", session.ID)
	}

	var liftIds = []primitive.ObjectID{}
	liftIds = append(liftIds, session.Lifts...)

	liftFilter := bson.M{"_id": bson.M{"$in": liftIds}}
	var lifts []Lift

	liftCursor, err := liftCollection.Find(context.TODO(), liftFilter)
	if err != nil {
		log.Fatal(err)
	}
	defer liftCursor.Close(context.TODO())

	for liftCursor.Next(context.TODO()) {
		var doc Lift
		err := liftCursor.Decode(&doc)
		if err != nil {
			log.Fatal(err)
		}
		lifts = append(lifts, doc)
	}

	var sessionResponse = SessionResponse{ID: session.ID, Floors: session.Floors, Lifts: lifts}

	return &sessionResponse, nil
}

func CreateLiftRequest(floor int, session_id string) (*LiftRequest, error) {
	var lift Lift
	liftFilters := bson.M{"status": StatusIdle}

	err := liftCollection.FindOne(context.TODO(), liftFilters).Decode(&lift)
	if err != nil {
		return nil, &utils.CustomError{Message: "No lift available right now"}
	}

	var session Session
	session_oid, err := primitive.ObjectIDFromHex(session_id)
	if err != nil {
		log.Fatal(err)
	}
	sessionFilters := bson.M{"_id": session_oid}

	err = sessionCollection.FindOne(context.TODO(), sessionFilters).Decode(&session)
	if err != nil {
		return nil, &utils.CustomError{Message: "Session Not Find"}
	}

	liftRequest := LiftRequest{RequestedFloor: floor, Status: StatusQueued, Lift: lift.ID, Session: session_oid}

	result, err := liftRequestCollection.InsertOne(context.TODO(), liftRequest)
	if err != nil {
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		liftRequest.ID = oid
	} else {
		return nil, fmt.Errorf("failed to retrieve inserted ID")
	}

	return &liftRequest, nil
}

func GetLiftRequests(session_id string, requestStatus string) ([]*LiftRequest, error) {
	if err := ValidateLiftRequestStatus(requestStatus); err != nil {
		return nil, &utils.CustomError{Message: requestStatus + " ,is not a valid status"}
	}

	liftRequestsFilters := bson.M{}

	if requestStatus != "" {
		liftRequestsFilters["status"] = requestStatus
	}

	if session_id != "" {
		var session Session

		session_oid, err := primitive.ObjectIDFromHex(session_id)
		if err != nil {
			log.Fatal(err)
		}

		sessionFilters := bson.M{"_id": session_oid}

		err = sessionCollection.FindOne(context.TODO(), sessionFilters).Decode(&session)
		if err != nil {
			fmt.Print(err, session_id)
			return nil, &utils.CustomError{Message: "Session Not Find"}
		}

		liftRequestsFilters["session"] = session_oid
	}

	curr, err := liftRequestCollection.Find(context.TODO(), liftRequestsFilters)
	if err != nil {
		return nil, err
	}

	var results []*LiftRequest
	for curr.Next(context.TODO()) {
		var result LiftRequest
		e := curr.Decode(&result)
		if e != nil {
			log.Fatal(e)
		}
		results = append(results, &result)
	}

	if err := curr.Err(); err != nil {
		return nil, err
	}

	defer curr.Close(context.TODO())
	return results, nil
}
