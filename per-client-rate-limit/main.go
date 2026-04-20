package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"net"

	"golang.org/x/time/rate"
)

// Message represents the standard JSON response format for the API.
type Message struct {
	Status string `json:"status"` // The status of the response (e.g., "Successful" or "Request failed").
	Body   string `json:"body"`   // The human-readable body of the response.
}

func perClientRateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc{
	
	// Define a custom struct to hold the rate limiter and the last time the client was seen.
	type client struct{
		limiter *rate.Limiter // The actual token bucket limiter for the specific client.
		lastSeen time.Time    // Timestamp of the last request made by this client.
	}

	var (
		mu sync.Mutex                      // Mutex to ensure thread-safe access to the clients map.
		clients = make(map[string]*client) // Map correlating client IP addresses to their specific rate limiter configuration.
	)

	// Launch a background goroutine to clean up inactive clients and free memory.
	go func(){
		for{
			time.Sleep(time.Minute) // Wait for one minute before running the cleanup task again.
			mu.Lock()               // Lock the mutex before reading or modifying the clients map.
			
			// Iterate over all tracked clients by their IP.
			for ip, client := range clients{
				// If the client hasn't made a request in over 3 minutes, remove them.
				if time.Since(client.lastSeen) > 3*time.Minute{
					delete(clients, ip) // Delete the inactive client from the map.
				}
			}
			mu.Unlock() // Unlock the mutex to allow other operations.
		}
	}()

	// Return an HTTP handler function that processes each incoming request.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		// Extract the IP address from the request's RemoteAddr (which is formatted as "IP:port").
		ip, _, err :=net.SplitHostPort(r.RemoteAddr)
		if err != nil{
			// If parsing fails, send a 500 status and stop.
			w.WriteHeader(http.StatusInternalServerError)
			return 	
		}
		
		mu.Lock() // Lock the map since it's going to be accessed concurrently by multiple incoming requests.
		
		// If the IP is entirely new and not currently tracked in the map...
		if _, found := clients[ip]; !found {
			// Initialize a new client entry with a limiter allowing 2 requests per second with a burst of 4.
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}	
		}
		
		// Update the specific client's last seen timestamp to the current time.
		clients[ip].lastSeen = time.Now()
		
		// Check if the specific client's token bucket allows another request.
		if !clients[ip].limiter.Allow(){
			mu.Unlock() // Crucial: Unlock the mutex before sending a response to prevent deadlocks.
			
			// Formulate an error message signifying capacity is exceeded.
			message := Message{
				Status: "Request failed",
				Body: "The Api is at capacity, try again later.",
			}
			w.WriteHeader(http.StatusTooManyRequests) // Send HTTP 429 (Too Many Requests) status code.
			json.NewEncoder(w).Encode(&message)       // Send the JSON response to the client.
			return
		} 
		
		mu.Unlock() // Unlock the map before processing the valid request.
		
		// Forward the request to the upstream/next handler.
		next(w, r)	
	})
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
	err:=json.NewEncoder(w).Encode(&message)
	if err != nil{
		return // If JSON encoding fails, stop execution silently.
	}
}

func main(){
	// Register the "/ping" route, but wrap the core endpointHandler with our custom perClientRateLimiter middleware.
	http.Handle("/ping", perClientRateLimiter(endpointHandler))
	
	// Start the standard HTTP server listening on port 8080.
	err := http.ListenAndServe(":8080", nil)
	if err != nil{
		// If the server fails to start, log the error to the console.
		log.Println("There was an error listening to port:8080", err)
	}
	
}