package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var llmClient LLMClient

	if os.Getenv("ENV") == "production" {
		log.Println("Running in production mode, using Replicate client")
		llmClient = NewReplicateClient(os.Getenv("REPLICATE_API_TOKEN"))
	} else {
		log.Println("Running in dev mode, using Ollama client")
		llmClient = NewOllamaClient(os.Getenv("LLM_URL"))
	}

	// Route handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "The AI Vibe DJ is listening...")
	})

	http.HandleFunc("/findTheVibe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Protect server from excessively large requests. 100MB limit.
		r.Body = http.MaxBytesReader(w, r.Body, 100*1024*1024)

		imageData, err := io.ReadAll(r.Body)
		if err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				http.Error(w, "request body is too large", http.StatusRequestEntityTooLarge)
				return
			}
			log.Printf("error reading request body: %v", err)
			http.Error(w, "could not read request body", http.StatusBadRequest)
			return
		}

		// TODO(dse): get prompt from the request and iterate here
		prompt := "what is the musical vibe of this image?"

		desc, err := llmClient.DescribeImage(r.Context(), imageData, prompt)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to describe image: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "The vibe is: %s", desc)
	})

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
