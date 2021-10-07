package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func handler(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic("unexpected error during UUID generation")
	}
	fmt.Fprintf(w, "Random UUID: %v", uuid)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
