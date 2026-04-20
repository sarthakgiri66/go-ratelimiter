package main

import (
	"encoding/json"
	"log"
	"net/http"

	tollbooth "github.com/didip/tollbooth/v6"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func endpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	message := Message{
		Status: "Successful",
		Body:   "Hello, You're reached the API. How may I help you?",
	}
	err := json.NewEncoder(w).Encode(&message)
	if err != nil {
		return
	}
}

func main() {
	message := Message{
		Status: "Request failed",
		Body:   "The Api is at capacity, try again later.",
	}
	jsonMessage, _ := json.Marshal(message)

	tlbthLimiter := tollbooth.NewLimiter(1, nil)
	tlbthLimiter.SetMessageContentType("application/json")
	tlbthLimiter.SetMessage(string(jsonMessage))

	http.Handle("/ping", tollbooth.LimitFuncHandler(tlbthLimiter, endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("There was an error listening to port:8080", err)
	}
}
