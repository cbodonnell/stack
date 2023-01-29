package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// add a flag to configure the port
	var port int
	// TODO: Fix flags
	flag.IntVar(&port, "port", 8080, "port to listen on")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", r.URL.Path)
	})

	log.Printf("Listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
