package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"./routes/project"
	"github.com/cjhammons/nommer/routes/project"
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

	fmt.Println("Connected to MongoDB!")
	// Create or get a collection
	collection := client.Database(mongoDatabase).Collection("project-events")

	router := mux.NewRouter()

	router.HandleFunc("/1/projects", &project.CreateProjectHandler(collection)).Methods("POST")
	router.HandleFunc("/1/projects/{projectname}/events", &project.CreateEventHandler(collection)).Methods("POST")

	// Wrap router with Gorilla Handlers for additional functionality like Logging
	loggingRouter := handlers.LoggingHandler(http.Stdout, router)

	// Start server
	srv := &http.Server{
		Handler:      loggingRouter,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
