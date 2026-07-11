package main

import (
	"fmt"
	"log"
	"net/http"
)

/*
1. send image to LLM
2. get description
3. auth with spotify client
4. search for spotify playlists (filter null)
5. get all tracks for spotify playlistIDs
6. sort tracks by popularity? Or just choose one?
7. return trackID
*/

func main() {
	spotifyClient := NewSpotifyClient()
	aiClient := FakeAIClient{}

	http.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "pong")
	})

	http.HandleFunc("/image-description", func(writer http.ResponseWriter, request *http.Request) {
		// TODO(dse): This will need to get tags from LLM
		imageTags := aiClient.ImageTags()

		trackID, err := spotifyClient.GetTrackIDFromDescription(imageTags)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(writer, trackID)
	})

	log.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

type FakeAIClient struct{}

func (client FakeAIClient) ImageTags() []string {
	return []string{"sunny", "beach", "sand", "palm trees", "relax"}
}
