package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Message represents the standard JSON response format for the API.
type Message struct {
	Status string `json:"status"` // The status of the response (e.g., "Successful" or "Request failed").
	Body   string `json:"body"`   // The human-readable body of the response.
}

// endpointHandler represents the core business logic of the API endpoint.
func endpointHandler(w http.ResponseWriter, r *http.Request){
	// Set the Content-Type header to explicitly indicate a JSON response.
	w.Header().Set("Content-Type", "application/json")
	// Set the HTTP status code to 200 OK for successful requests.
	w.WriteHeader(http.StatusOK)
	
	// Create the success message object.
	message := Message{
		Status: "Successful",
		Body: "Hello, You're reached the API. How may I help you?",
	}
	
	// Dynamically encode and send the message struct as a JSON response directly into the ResponseWriter.
	err := json.NewEncoder(w).Encode(&message)
	if err != nil{
		return // If JSON encoding fails, stop execution silently.
	}
}

func main(){
	// Register the "/ping" route, but wrap the core endpointHandler with our global token rateLimiter middleware.
	http.Handle("/ping", rateLimiter(endpointHandler))
	
	// Start the standard HTTP server listening on port 8080.
	err := http.ListenAndServe(":8080", nil)
	if err != nil{
		// If the server fails to start (e.g. port already in use), log the error to the console.
		log.Println("There was an error listening to port:8080")
	}	
}