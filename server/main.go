package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var llmClient LLMClient
	var spotifyClient *SpotifyClient

	if os.Getenv("ENV") == "production" {
		log.Println("Running in production mode, using Replicate client")
		llmClient = NewReplicateClient(os.Getenv("REPLICATE_API_TOKEN"))
	} else {
		log.Println("Running in dev mode, using Ollama client")
		llmClient = NewOllamaClient(os.Getenv("LLM_URL"))
	}

	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if spotifyClientID == "" || spotifyClientSecret == "" {
		log.Println("WARNING: SPOTIFY_CLIENT_ID or SPOTIFY_CLIENT_SECRET not set. Spotify API calls will fail.")
	}
	spotifyClient = NewSpotifyClient(spotifyClientID, spotifyClientSecret)

	// Fetch genre seeds once on startup
	log.Println("Fetching Spotify genre seeds...")
	availableGenres, err := spotifyClient.getAvailableGenreSeeds(context.Background())
	if err != nil {
		log.Fatalf("FATAL: could not fetch spotify genre seeds on startup: %v", err)
	}
	log.Printf("Successfully fetched %d Spotify genre seeds.", len(availableGenres))
	genreList := strings.Join(availableGenres, ", ")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "The AI Vibe DJ is listening...")
	})

	http.HandleFunc("/findTheVibe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

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

		prompt := fmt.Sprintf(`Analyze the following image and describe its musical vibe. 
Based on your description, choose up to 5 genres from the following list that best match the vibe.
Available genres: %s
Return ONLY a JSON object with two keys: "description" (a string) and "genres" (an array of strings).`, genreList)

		desc, err := llmClient.DescribeImage(r.Context(), imageData, prompt)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to describe image: %v", err), http.StatusInternalServerError)
			return
		}

		// TODO(dse): get song recommendation from spotify
		log.Printf("LLM described vibe as: %s", desc)
		track, err := spotifyClient.GetSongRecommendation(r.Context(), desc)
		if err != nil {
			log.Printf("ERROR: failed to get spotify recommendation: %v", err)
			http.Error(w, "failed to get spotify recommendation", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "The recommended track is: %s by %s", track.Name, track.Artists[0].Name)
	})

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
