package main

import (
	"log"
	"os"
)

type Client interface {
	ready() bool
}

type SpotifyClient struct {
	clientID string
}

func (SpotifyClient) ready() bool {
	return true
}

func NewSpotifyClient() SpotifyClient {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	if clientID == "" {
		log.Fatal("missing SPOTIFY_CLIENT_ID")
	}

	return SpotifyClient{clientID: clientID}
}
