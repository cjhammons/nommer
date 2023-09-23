package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Project struct {
	ID     string  `json:"_id" bson:"_id,omitempty"`
	Name   string  `json:"name" bson:"name"`
	APIKey string  `json:"apikey" bson:"apikey"`
	Events []Event `json:"events" bson:"events"`
}

type Event struct {
	TimeStamp time.Time              `json:"timestamp" bson:"timestamp"`
	Data      map[string]interface{} `json:"Event" bson:"Event"`
}

// CreateProjectHandler creates a new project and assigns an API key
func CreateProjectHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

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
		log.Println("Checking if project {name} already exists", projectName)
		err := collection.FindOne(ctx, bson.M{"name": projectName}).Decode(&existingProject)
		if err == nil {
			http.Error(w, "Project name already exists", http.StatusConflict)
			return
		}

		// Create a Project object
		p := Project{
			Name:   projectName,
			APIKey: GenerateAPIKey(projectName),
			Events: []Event{},
		}
		// Insert the project into the database
		_, err = collection.InsertOne(ctx, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Created project {name}", projectName)
		// Return the created project, including the API key
		json.NewEncoder(w).Encode(p)
	}
}

func SendProjectEventHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		vars := mux.Vars(r)
		projectName := vars["project_name"]

		// Retrieve API key from request header
		apiKey := r.Header.Get("X-API-Key")

		// Find project by name and validate API key
		var foundProject Project
		err := collection.FindOne(ctx, bson.M{"name": projectName}).Decode(&foundProject)
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
		_, err = collection.UpdateOne(ctx, bson.M{"name": projectName}, update)
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
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}
	data := fmt.Sprintf("%s-%d-%d", projectName, time.Now().UnixNano(), n)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
