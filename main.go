package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	// Setup MongoDB connection
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
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
	collection := client.Database("mydatabase").Collection("projects")

	router := mux.NewRouter()

	router.HandleFunc("/1/projects", CreateProjectHandler(collection)).Methods("POST")
	router.HandleFunc("/1/projects/{projectname}/events", CreateEventHandler(collection)).Methods("POST")

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
