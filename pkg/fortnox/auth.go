// Package fortnox provides functionality for authenticating with the Fortnox API.
package fortnox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	authorizationEndpoint = "https://apps.fortnox.se/oauth-v1/auth"
	tokenEndpoint         = "https://apps.fortnox.se/oauth-v1/token"
)

type FortnoxClient struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       string
	APIBaseURL   string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	AuthDone     chan bool
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func NewFortnoxClient() (*FortnoxClient, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	client := &FortnoxClient{
		ClientID:     os.Getenv("FORTNOX_CLIENT_ID"),
		ClientSecret: os.Getenv("FORTNOX_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("REDIRECT_URI"),
		Scopes:       os.Getenv("FORTNOX_CLIENT_SCOPES"),
		APIBaseURL:   os.Getenv("FORTNOX_API_BASE_URL"),
		AuthDone:     make(chan bool),
	}

	err = client.loadTokens()
	if err != nil {
		// It's okay if we can't load tokens, we might need to get new ones
		fmt.Println("No saved tokens found. New authorization might be required.")
	}

	return client, nil
}

func (c *FortnoxClient) GetAuthorizationURL(state string) string {
	return fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=%s&state=%s&response_type=code&access_type=offline",
		authorizationEndpoint, c.ClientID, c.RedirectURI, c.Scopes, state)
}

func (c *FortnoxClient) ExchangeAuthorizationCode(code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.RedirectURI)

	return c.doTokenRequest(data)
}

func (c *FortnoxClient) RefreshAccessToken() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.RefreshToken)

	return c.doTokenRequest(data)
}

func (c *FortnoxClient) doTokenRequest(data url.Values) error {
	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.ClientID, c.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	c.AccessToken = tokenResp.AccessToken
	c.RefreshToken = tokenResp.RefreshToken
	c.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return c.saveTokens()
}

func (c *FortnoxClient) saveTokens() error {
	data, err := json.Marshal(map[string]string{
		"access_token":  c.AccessToken,
		"refresh_token": c.RefreshToken,
		"expires_at":    c.ExpiresAt.Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	return os.WriteFile("fortnox_tokens.json", data, 0600)
}

func (c *FortnoxClient) loadTokens() error {
	data, err := os.ReadFile("fortnox_tokens.json")
	if err != nil {
		return err
	}

	var tokens map[string]string
	if err := json.Unmarshal(data, &tokens); err != nil {
		return err
	}

	c.AccessToken = tokens["access_token"]
	c.RefreshToken = tokens["refresh_token"]
	c.ExpiresAt, _ = time.Parse(time.RFC3339, tokens["expires_at"])

	return nil
}

func (c *FortnoxClient) StartAuthorizationFlow() error {
	state := "some_random_state" // Change this to a more secure state if needed
	authURL := c.GetAuthorizationURL(state)

	// Open the browser with the authorization URL
	err := exec.Command("open", authURL).Start()
	if err != nil {
		return fmt.Errorf("failed to open browser: %v", err)
	}

	// Parse the redirect URI to extract the path
	parsedRedirectURI, err := url.Parse(c.RedirectURI)
	if err != nil {
		return fmt.Errorf("invalid redirect URI: %v", err)
	}

	// Start a simple HTTP server to listen for the callback
	server := &http.Server{Addr: parsedRedirectURI.Host}

	http.HandleFunc(parsedRedirectURI.Path, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		code := query.Get("code")
		stateReceived := query.Get("state")

		if state != stateReceived {
			http.Error(w, "state does not match", http.StatusBadRequest)
			return
		}

		err := c.ExchangeAuthorizationCode(code)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to exchange authorization code: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Authorization successful! You can close this window.")
		c.AuthDone <- true
	})

	fmt.Printf("Listening on %s for the authorization code...\n", c.RedirectURI)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", c.RedirectURI, err)
		}
	}()

	<-c.AuthDone
	fmt.Println("Authorization successful, signal authDone channel")

	// Create a context with a timeout to shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server Shutdown Failed:%+v", err)
	}

	return nil
}
