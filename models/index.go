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
	ID     primitive.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	Lifts  []Lift             `json:"lifts"`
	Floors int                `json:"floors"`
}

type SessionDocument struct {
	ID     primitive.ObjectID   `json:"_id,omitempty"  bson:"_id,omitempty"`
	Lifts  []primitive.ObjectID `json:"lifts"`
	Floors int                  `json:"floors"`
}

type LiftRequest struct {
	ID             primitive.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	RequestedFloor int                `json:"requestedFloor"`
	Lift           primitive.ObjectID `json:"lift"`
	Status         string             `json:"status,omitempty"`
	Session        primitive.ObjectID `json:"session"`
}

type LiftRequestResponse struct {
	ID             primitive.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	RequestedFloor int                `json:"requestedFloor"`
	Lift           Lift               `json:"lift"`
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
		log.Fatal("Error loading the ENV file")
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
		log.Fatalf("Error connecting to MongoDB: %v", err)
		return
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
		return
	}

	fmt.Println("Connected to MongoDB")

	liftCollection = client.Database(dbName).Collection(liftCollName)
	liftRequestCollection = client.Database(dbName).Collection(liftRequestCollName)
	sessionCollection = client.Database(dbName).Collection(sessionCollName)
}

func CreateSession(floors int, lifts int) (*Session, error) {
	var interfacesObjs []interface{}
	for i := 0; i < lifts; i++ {
		lift := Lift{CurrentFloor: 0, Status: "idle"}
		interfacesObjs = append(interfacesObjs, lift)
	}

	liftResults, err := liftCollection.InsertMany(context.Background(), interfacesObjs)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Create a slice to store the inserted lift IDs.
	var insertedLiftIDs []primitive.ObjectID
	for _, result := range liftResults.InsertedIDs {
		insertedLiftIDs = append(insertedLiftIDs, result.(primitive.ObjectID))
	}

	var insertedLifts []Lift
	for _, id := range insertedLiftIDs {
		insertedLifts = append(insertedLifts, Lift{ID: id, CurrentFloor: 0, Status: "idle"})
	}

	// Create the Session object with the inserted lift IDs.
	sessionDoc := SessionDocument{Floors: floors, Lifts: insertedLiftIDs}
	result, err := sessionCollection.InsertOne(context.Background(), sessionDoc)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	session := Session{Floors: floors, Lifts: insertedLifts, ID: result.InsertedID.(primitive.ObjectID)}
	return &session, nil
}

func GetSession(sessionID string) (*Session, error) {
	sessionObjectID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	sessionFilter := bson.M{"_id": sessionObjectID}
	var sessionDoc SessionDocument

	err = sessionCollection.FindOne(context.TODO(), sessionFilter).Decode(&sessionDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Session Document not found")
		} else {
			fmt.Println(err)
		}
		return nil, err
	} else {
		fmt.Printf("Found document: %+v\n", sessionDoc.ID)
	}

	var liftIds = []primitive.ObjectID{}
	liftIds = append(liftIds, sessionDoc.Lifts...)

	liftFilter := bson.M{"_id": bson.M{"$in": liftIds}}
	var lifts []Lift

	liftCursor, err := liftCollection.Find(context.TODO(), liftFilter)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer liftCursor.Close(context.TODO())

	for liftCursor.Next(context.TODO()) {
		var doc Lift
		err := liftCursor.Decode(&doc)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		lifts = append(lifts, doc)
	}

	session := Session{ID: sessionDoc.ID, Floors: sessionDoc.Floors, Lifts: lifts}

	return &session, nil
}

func CreateLiftRequest(floor int, sessionID string) (*LiftRequest, *LiftRequestResponse, error) {
	sessionObjectID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	sessionFilter := bson.M{"_id": sessionObjectID}
	var sessionDoc SessionDocument

	err = sessionCollection.FindOne(context.TODO(), sessionFilter).Decode(&sessionDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Session Document not found")
		} else {
			fmt.Println(err)
		}
		return nil, nil, err
	} else {
		fmt.Printf("Found document Session: %+v\n", sessionDoc.ID)
	}

	liftRequestFilter := bson.M{"session": sessionObjectID, "requestedfloor": floor, "status": StatusQueued}
	var liftRequestPresent LiftRequest
	err = liftRequestCollection.FindOne(context.TODO(), liftRequestFilter).Decode(&liftRequestPresent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Lift Request Document not found")
		} else {
			fmt.Println(err)
		}
		return nil, nil, err
	} else {
		fmt.Printf("Found document Lift Request: %+v\n", liftRequestPresent.ID)
	}

	if liftRequestPresent.ID != primitive.NilObjectID {
		return nil, nil, &utils.CustomError{Message: "Already a lift is called for the floor"}
	}

	var liftIds = []primitive.ObjectID{}
	liftIds = append(liftIds, sessionDoc.Lifts...)

	liftFilter := bson.M{"_id": bson.M{"$in": liftIds}, "status": StatusIdle}
	var lift Lift

	err = liftCollection.FindOne(context.TODO(), liftFilter).Decode(&lift)
	if err != nil {
		return nil, nil, &utils.CustomError{Message: "No lift available right now"}
	}

	liftRequest := LiftRequest{RequestedFloor: floor, Status: StatusQueued, Lift: lift.ID, Session: sessionObjectID}

	result, err := liftRequestCollection.InsertOne(context.TODO(), liftRequest)
	if err != nil {
		return nil, nil, err
	}
	liftRequest.ID = result.InsertedID.(primitive.ObjectID)

	_, err = liftCollection.UpdateOne(context.TODO(), bson.M{"_id": lift.ID}, bson.M{"$set": bson.M{
		"status": "busy",
	}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Lift Document not found")
		} else {
			fmt.Println(err)
		}
		return nil, nil, err
	} else {
		fmt.Printf("Update document successfully: %+v\n", lift.ID.Hex())
	}

	return &liftRequest, &LiftRequestResponse{ID: result.InsertedID.(primitive.ObjectID), RequestedFloor: floor, Status: StatusQueued, Lift: lift, Session: sessionObjectID}, nil
}

func GetLiftRequests(sessionID string, requestStatus string) ([]*LiftRequest, error) {
	if requestStatus == "" {
		requestStatus = StatusQueued
	}
	if err := ValidateLiftRequestStatus(requestStatus); err != nil {
		return nil, &utils.CustomError{Message: requestStatus + " ,is not a valid status"}
	}

	liftRequestsFilters := bson.M{}
	liftRequestsFilters["status"] = requestStatus

	if sessionID != "" {
		var sessionDoc SessionDocument
		sessionObjectID, err := primitive.ObjectIDFromHex(sessionID)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		sessionFilters := bson.M{"_id": sessionObjectID}
		err = sessionCollection.FindOne(context.TODO(), sessionFilters).Decode(&sessionDoc)
		if err != nil {
			fmt.Println(err, sessionID)
			return nil, &utils.CustomError{Message: "Session Not Found"}
		}

		liftRequestsFilters["session"] = sessionObjectID
	}

	curr, err := liftRequestCollection.Find(context.TODO(), liftRequestsFilters)
	if err != nil {
		return nil, err
	}
	defer curr.Close(context.TODO())

	var results []*LiftRequest
	for curr.Next(context.TODO()) {
		var result LiftRequest
		e := curr.Decode(&result)
		if e != nil {
			fmt.Println(err)
		}
		results = append(results, &result)
	}

	if err := curr.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func CompleteLiftRequest(liftRequest *LiftRequest) error {
	liftRequestFilter := bson.M{"_id": liftRequest.ID}
	updatedLiftRequest := bson.M{"$set": bson.M{
		"status": "completed",
	}}

	_, err := liftRequestCollection.UpdateOne(context.TODO(), liftRequestFilter, updatedLiftRequest)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Lift Request Document not found")
		} else {
			fmt.Println(err)
		}
		return err
	} else {
		fmt.Printf("Update document successfully LiftRequest: %+v\n", liftRequest.ID.Hex())
	}

	_, err = liftCollection.UpdateOne(context.TODO(), bson.M{"_id": liftRequest.Lift}, bson.M{"$set": bson.M{
		"status": "idle", "currentfloor": liftRequest.RequestedFloor,
	}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Lift Document not found")
		} else {
			fmt.Println(err)
		}
		return err
	} else {
		fmt.Printf("Update document successfully Lift: %+v\n", liftRequest.Lift.Hex())
	}

	return nil
}
