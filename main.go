package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cjhammons/nommer/routes"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Connecting to mongoDB")
	// Read environment variables
	mongoURI := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DATABASE")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Setup MongoDB connection
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB:" + mongoURI + mongoDatabase)
	// Create or get a collection
	collection := client.Database(mongoDatabase).Collection("project-events")

	router := mux.NewRouter()

	router.HandleFunc("/1/projects", routes.CreateProjectHandler(collection)).Methods("POST")
	router.HandleFunc("/1/projects", routes.GetProjectsHandler(collection)).Methods("GET")
	router.HandleFunc("/1/{project_name}/event", routes.SendProjectEventHandler(collection)).Methods("POST")

	// Wrap router with Gorilla Handlers for additional functionality like Logging
	loggingRouter := handlers.LoggingHandler(os.Stdout, router)

	// Start server
	srv := &http.Server{
		Handler:      loggingRouter,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting server on port 8080")

	log.Fatal(srv.ListenAndServe())
}
