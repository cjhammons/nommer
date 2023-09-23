package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Project struct {
	ID     string  `json:"id" bson:"_id,omitempty"`
	Name   string  `json:"name" bson:"name"`
	APIKey string  `json:"apikey" bson:"apikey"`
	Events []Event `json:"events" bson:"events"`
}

type Event struct {
	TimeStamp time.Time              `json:"type" bson:"type"`
	Data      map[string]interface{} `json:"data" bson:"data"`
}

// CreateProjectHandler creates a new project and assigns an API key
func CreateProjectHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Extract project name from payload
		projectName, exists := payload["name"]
		if !exists {
			http.Error(w, "Project name is required", http.StatusBadRequest)
			return
		}

		// Check if project name already exists
		var existingProject Project
		err := collection.FindOne(context.TODO(), bson.M{"name": projectName}).Decode(&existingProject)
		if err == nil {
			http.Error(w, "Project name already exists", http.StatusConflict)
			return
		}

		// Create a Project object
		var p Project
		p.Name = projectName

		// Generate an API key for the project
		p.APIKey = GenerateAPIKey(projectName)

		// Insert the project into the database
		_, err = collection.InsertOne(context.TODO(), p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the created project, including the API key
		json.NewEncoder(w).Encode(p)
	}
}

func SendProjectEventHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		projectName := vars["projectName"]

		// Retrieve API key from request header
		apiKey := r.Header.Get("X-API-Key")

		// Find project by name and validate API key
		var foundProject Project
		err := collection.FindOne(context.TODO(), bson.M{"name": projectName}).Decode(&foundProject)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		if foundProject.APIKey != apiKey {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// Decode the incoming JSON into a map
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if 'event' key exists
		eventData, exists := payload["event"]
		if !exists {
			http.Error(w, "Event data is required", http.StatusBadRequest)
			return
		}

		// Create the Event object
		e := Event{
			Data:      eventData.(map[string]interface{}),
			TimeStamp: time.Now(),
		}

		// Push the event into MongoDB
		update := bson.D{{"$push", bson.D{{"events", e}}}}
		_, err = collection.UpdateOne(context.TODO(), bson.M{"name": projectName}, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the created event
		json.NewEncoder(w).Encode(e)
	}
}

func GenerateAPIKey(projectName string) string {
	max := big.NewInt(1<<31 - 1)
	data := fmt.Sprintf("%s-%d-%d", projectName, time.Now().UnixNano(), rand.Int(rand.Reader, max))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
