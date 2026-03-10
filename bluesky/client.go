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

// Facet represents a rich text annotation in a Bluesky post.
type Facet struct {
	Index    FacetIndex     `json:"index"`
	Features []FacetFeature `json:"features"`
}

type FacetIndex struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}

type FacetFeature struct {
	Type string `json:"$type"`
	Tag  string `json:"tag"`
}

// extractHashtagFacets finds all #hashtags in text and returns facets with byte offsets.
func extractHashtagFacets(text string) []Facet {
	var facets []Facet
	b := []byte(text)
	i := 0
	for i < len(b) {
		if b[i] == '#' {
			start := i
			i++
			tagStart := i
			for i < len(b) && (isLetter(b[i]) || isDigit(b[i]) || b[i] == '_') {
				i++
			}
			if i > tagStart {
				tag := string(b[tagStart:i])
				facets = append(facets, Facet{
					Index: FacetIndex{ByteStart: start, ByteEnd: i},
					Features: []FacetFeature{
						{Type: "app.bsky.richtext.facet#tag", Tag: tag},
					},
				})
			}
		} else {
			i++
		}
	}
	return facets
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// PostText creates a new text post on Bluesky, with hashtag facets automatically detected.
func (c *Client) PostText(text string) error {
	facets := extractHashtagFacets(text)

	record := map[string]any{
		"$type":     "app.bsky.feed.post",
		"text":      text,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}
	if len(facets) > 0 {
		record["facets"] = facets
	}

	return c.post("com.atproto.repo.createRecord", map[string]any{
		"repo":       c.did,
		"collection": "app.bsky.feed.post",
		"record":     record,
	}, nil)
}
