package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Project struct {
	ID     string  `json:"id" bson:"_id,omitempty"`
	Name   string  `json:"name" bson:"name"`
	APIKEY string  `json:"apikey" bson:"apikey"`
	Events []Event `json:"events" bson:"events"`
}

type Event struct {
	Type string                 `json:"type" bson:"type"`
	Data map[string]interface{} `json:"data" bson:"data"`
}

func CreateProjectHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var project Project
		err := json.NewDecoder(r.Body).Decode(&project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		insertResult, err := collection.InsertOne(context.Background(), project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(insertResult)
	}
}

func SendProjectEventHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var event Event
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		filter := bson.D{{"_id", vars["id"]}}
		update := bson.D{{"$push", bson.D{{"events", event}}}}
		updateResult, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(updateResult)
	}
}
