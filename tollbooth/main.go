package main

import (
	"encoding/json"
	"log"
	"net/http"

	tollbooth "github.com/didip/tollbooth/v6"
)

// Message represents the standard JSON response format for the API.
type Message struct {
	Status string `json:"status"` // The status of the response (e.g., "Successful" or "Request failed").
	Body   string `json:"body"`   // The human-readable body of the response.
}

// endpointHandler represents the core business logic of the API endpoint.
func endpointHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header to explicitly indicate a JSON response.
	w.Header().Set("Content-Type", "application/json")
	// Set the HTTP status code to 200 OK for successful requests.
	w.WriteHeader(http.StatusOK)
	
	// Create the success message object.
	message := Message{
		Status: "Successful",
		Body:   "Hello, You're reached the API. How may I help you?",
	}
	
	// Dynamically encode and send the message struct as a JSON response directly into the ResponseWriter.
	err := json.NewEncoder(w).Encode(&message)
	if err != nil {
		return // If JSON encoding fails, stop execution silently.
	}
}

func main() {
	// Define the message that will be sent back when a user exceeds the rate limit.
	message := Message{
		Status: "Request failed",
		Body:   "The Api is at capacity, try again later.",
	}
	// Pre-marshal the rate-limit block message into a JSON string format to reuse it for all blocked requests.
	jsonMessage, _ := json.Marshal(message)

	// Create a new Tollbooth rate limiter. The '1' signifies 1 request per second.
	tlbthLimiter := tollbooth.NewLimiter(1, nil)
	// Tell the Tollbooth limiter to respond with "application/json" on rejection.
	tlbthLimiter.SetMessageContentType("application/json")
	// Set the rejected response payload to our custom predefined JSON string.
	tlbthLimiter.SetMessage(string(jsonMessage))

	// Register the "/ping" route, wrapping our endpointHandler in Tollbooth's LimitFuncHandler middleware.
	http.Handle("/ping", tollbooth.LimitFuncHandler(tlbthLimiter, endpointHandler))
	
	// Start the standard HTTP server listening on port 8080.
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		// Log any error that prevents the server from starting properly.
		log.Println("There was an error listening to port:8080", err)
	}
}
