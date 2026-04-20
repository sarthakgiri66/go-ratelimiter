package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/time/rate"
)

func rateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler{
	// Create a new token bucket rate limiter allowing 2 requests per second with a burst of up to 4 requests.
	limiter := rate.NewLimiter(2, 4)
	
	// Return an HTTP handler that intercepts incoming requests before passing them to the next handler.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		// Check if a token is available in the bucket to allow the request to proceed.
		if !limiter.Allow(){
			// If no token is available, format a JSON error message object.
			message := Message{
				Status: "Request failed",
				Body: "The Api is at capacity, try again later.",
			}
			// Set the HTTP status code to 429 Too Many Requests to signify rate limiting.
			w.WriteHeader(http.StatusTooManyRequests)
			// Encode and send the error message as a JSON response to the client.
			json.NewEncoder(w).Encode(&message)
			// Stop further execution and do not call the next handler.
			return
		} else{
			// If a token is available, pass the request and response objects to the next handler in the chain.
			next(w, r)
		}

})
}
