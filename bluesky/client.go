package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const pdsURL = "https://bsky.social/xrpc"

type Client struct {
	http        *http.Client
	handle      string
	appPassword string
	accessJWT   string
	did         string
}

func NewClient(handle, appPassword string) *Client {
	return &Client{
		http:        &http.Client{Timeout: 10 * time.Second},
		handle:      handle,
		appPassword: appPassword,
	}
}

func (c *Client) post(endpoint string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", pdsURL, endpoint), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.accessJWT != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessJWT)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("bluesky %s returned %d: %s", endpoint, resp.StatusCode, string(data))
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

// Login authenticates and stores the session JWT.
func (c *Client) Login() error {
	var result struct {
		AccessJWT string `json:"accessJwt"`
		DID       string `json:"did"`
	}
	err := c.post("com.atproto.server.createSession", map[string]string{
		"identifier": c.handle,
		"password":   c.appPassword,
	}, &result)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	c.accessJWT = result.AccessJWT
	c.did = result.DID
	return nil
}

// PostText creates a new text post on Bluesky.
func (c *Client) PostText(text string) error {
	return c.post("com.atproto.repo.createRecord", map[string]any{
		"repo":       c.did,
		"collection": "app.bsky.feed.post",
		"record": map[string]any{
			"$type":     "app.bsky.feed.post",
			"text":      text,
			"createdAt": time.Now().UTC().Format(time.RFC3339),
		},
	}, nil)
}
