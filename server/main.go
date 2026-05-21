package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "pong")
	})

	http.HandleFunc("/image-description", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "image description")
	})

	log.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
