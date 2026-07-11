package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type SpotifyClient interface {
	ensureToken() error
	getRandomPlaylistID([]string) (string, error)
	getRandomTrackID(string) (string, error)
	GetTrackIDFromDescription([]string) (string, error)
}

type spotifyClient struct {
	clientID       string
	token          string
	tokenExpiresAt time.Time
}

type playlistItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	Tracks      struct {
		Href  string `json:"href"`
		Total int    `json:"total"`
	} `json:"tracks"`
}
type spotifyGetPlaylistResponse struct {
	Playlists struct {
		Items []*playlistItem `json:"items"`
		Total int             `json:"total"`
		Next  string          `json:"next"`
	} `json:"playlists"`
}

type playlistTrackItem struct {
	Item struct {
		ID string `json:"id"`
	} `json:"item"`
}

type spotifyGetPlaylistTrackResponse struct {
	Items []*playlistTrackItem `json:"items"`
}

const (
	tokenURL    = "https://accounts.spotify.com/api/token"
	searchURL   = "https://api.spotify.com/v1/search"
	playlistURL = "https://api.spotify.com/v1/playlists" // then /<playlistID>/items?limit=10&market=US
)

func (client *spotifyClient) ensureToken() error {
	refreshTime := client.tokenExpiresAt.Add(-5 * time.Minute)
	if client.token != "" && time.Now().Before(refreshTime) {
		return nil
	}

	body := strings.NewReader("grant_type=client_credentials")

	req, err := http.NewRequest(http.MethodPost, tokenURL, body)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(client.clientID, os.Getenv("SPOTIFY_CLIENT_SECRET"))

	// this is the actual request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read the body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Spotify token: %s: %s", resp.Status, string(respBody))
	}

	// unmarshal
	var respJson struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(respBody, &respJson); err != nil {
		return err
	}

	client.token = respJson.AccessToken
	client.tokenExpiresAt = time.Now().Add(time.Duration(respJson.ExpiresIn) * time.Second)

	return nil
}

func (client *spotifyClient) getRandomPlaylistID(imageTags []string) (string, error) {
	query := strings.Join(imageTags, " ")
	params := url.Values{}
	params.Set("q", query)
	params.Set("type", "playlist")
	params.Set("limit", "10")
	params.Set("market", "US")

	searchRequestURL := searchURL + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, searchRequestURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+client.token)
	req.Header.Set("Accept", "application/json")

	// actual request is here
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err

	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var searchResponse spotifyGetPlaylistResponse
	if err := json.Unmarshal(respBody, &searchResponse); err != nil {
		return "", err
	}

	playlistItems := searchResponse.Playlists.Items
	var nonNilItems []*playlistItem
	for _, item := range playlistItems {
		if item != nil && item.Tracks.Total > 0 && item.ID != "" {
			nonNilItems = append(nonNilItems, item)
		}
	}

	randomPlaylist, err := choose(playlistItems)
	if err != nil {
		return "", err
	}

	return randomPlaylist.ID, nil
}

func (client *spotifyClient) getRandomTrackID(playlistID string) (string, error) {
	fmt.Println("getting random trackID now. received random playlistID", playlistID)
	// set query and params
	params := url.Values{}
	params.Set("limit", "10")
	params.Set("market", "US")

	// build req
	searchRequestURL := fmt.Sprintf("%s/%s/tracks?%s", playlistURL, playlistID, params.Encode())
	req, err := http.NewRequest(http.MethodGet, searchRequestURL, nil)
	if err != nil {
		return "", err
	}

	// set headers
	req.Header.Set("Authorization", "Bearer "+client.token)
	req.Header.Set("Accept", "application/json")

	// do
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	// remember to defer close
	defer resp.Body.Close()

	// read stream
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}

	// unmarshal response
	var getPlaylistTracksResponse spotifyGetPlaylistTrackResponse
	if err := json.Unmarshal(respBody, &getPlaylistTracksResponse); err != nil {
		return "", err
	}
	items := getPlaylistTracksResponse.Items

	// choose a trackID
	randomTrack, err := choose(items)

	return randomTrack.Item.ID, nil
}

func (client *spotifyClient) GetTrackIDFromDescription(imageTags []string) (string, error) {
	err := client.ensureToken()
	if err != nil {
		return "", err
	}

	// first get random playlist from search
	randomPlaylistID, err := client.getRandomPlaylistID(imageTags)
	if err != nil {
		return "", err
	}

	// Now get a random track from the playlist
	randomTrackID, err := client.getRandomTrackID(randomPlaylistID)
	if err != nil {
		return "", err
	}

	return randomTrackID, nil
}

func NewSpotifyClient() SpotifyClient {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	if clientID == "" {
		log.Fatal("missing SPOTIFY_CLIENT_ID")
	}

	return &spotifyClient{clientID: clientID}
}
