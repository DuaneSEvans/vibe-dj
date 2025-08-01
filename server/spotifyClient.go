package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type spotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type spotifyArtist struct {
	Name string `json:"name"`
}

type spotifyGenresResponse struct {
	Genres []string `json:"genres"`
}

// spotifyTrack is a simplified representation of a Spotify track.
type spotifyTrack struct {
	Name    string          `json:"name"`
	Artists []spotifyArtist `json:"artists"`
	URL     string          `json:"external_urls.spotify"`
}

type SpotifyClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	token        *spotifyTokenResponse
	tokenMu      sync.Mutex
	tokenExpiry  time.Time
}

func NewSpotifyClient(id, secret string) *SpotifyClient {
	return &SpotifyClient{
		clientID:     id,
		clientSecret: secret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *SpotifyClient) authenticate(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// If token is still valid, do nothing.
	if c.token != nil && time.Now().Before(c.tokenExpiry) {
		log.Println("[spotify] Token is still valid, skipping authentication")
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create spotify auth request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get spotify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("spotify auth request failed with status: %s", resp.Status)
	}

	var tokenResp spotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode spotify token response: %w", err)
	}

	c.token = &tokenResp
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

func (c *SpotifyClient) getAvailableGenreSeeds(ctx context.Context) ([]string, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("spotify authentication failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/recommendations/available-genre-seeds", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create genre seeds request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get genre seeds: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("genre seeds request failed with status: %s", resp.Status)
	}

	var genresResp spotifyGenresResponse
	if err := json.NewDecoder(resp.Body).Decode(&genresResp); err != nil {
		return nil, fmt.Errorf("failed to decode genre seeds response: %w", err)
	}

	return genresResp.Genres, nil
}

type llmVibeResponse struct {
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
}

type spotifyRecommendationResponse struct {
	Tracks []spotifyTrack `json:"tracks"`
}

// GetSongRecommendation takes a JSON string from the LLM and returns a recommended song.
func (c *SpotifyClient) GetSongRecommendation(ctx context.Context, llmJSONResponse string) (*spotifyTrack, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("spotify authentication failed: %w", err)
	}

	var vibe llmVibeResponse
	if err := json.Unmarshal([]byte(llmJSONResponse), &vibe); err != nil {
		return nil, fmt.Errorf("failed to unmarshal llm response: %w", err)
	}

	if len(vibe.Genres) == 0 {
		return nil, fmt.Errorf("llm response contained no genres")
	}

	log.Printf("[spotify] getting recommendation for genres: %v", vibe.Genres)

	// Build the request to Spotify's recommendation endpoint
	u, err := url.Parse("https://api.spotify.com/v1/recommendations")
	if err != nil {
		return nil, fmt.Errorf("failed to parse spotify recommendations url: %w", err)
	}

	q := u.Query()
	q.Set("limit", "5") // Get a few options
	q.Set("seed_genres", strings.Join(vibe.Genres, ","))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create spotify recommendation request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get spotify recommendation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify recommendation request failed with status: %s", resp.Status)
	}

	var recommendations spotifyRecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&recommendations); err != nil {
		return nil, fmt.Errorf("failed to decode spotify recommendation response: %w", err)
	}

	if len(recommendations.Tracks) == 0 {
		return nil, fmt.Errorf("spotify returned no tracks for the given genres")
	}

	// Return the first track from the list
	return &recommendations.Tracks[0], nil
}
