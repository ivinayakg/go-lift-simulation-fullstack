package utils

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomError struct {
	Message string
}

// Error returns the error message for the CustomError type.
func (e *CustomError) Error() string {
	return e.Message
}

func GenerateUUID() primitive.ObjectID {
	u := uuid.New()
	return primitive.ObjectID(u[:])
}
