package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	addr := ":4179"
	r := router()
	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Println(fmt.Sprintf("Listening and serving on %s", addr))
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Listening and serving failed: %v", err)
	}
}
