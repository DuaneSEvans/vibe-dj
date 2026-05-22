package main

import (
	"log"
	"os"
)

type SpotifyClient interface {
	ready() bool
}

type spotifyClient struct {
	clientID string
}

func (spotifyClient) ready() bool {
	return true
}

func NewSpotifyClient() SpotifyClient {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	if clientID == "" {
		log.Fatal("missing SPOTIFY_CLIENT_ID")
	}

	return spotifyClient{clientID: clientID}
}
