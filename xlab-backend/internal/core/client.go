package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Envelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) CurrentUser(ctx context.Context, token string) (*User, error) {
	var user User
	if err := c.getEnvelope(ctx, token, "/user/profile", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.getEnvelope(ctx, token, path, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (c *Client) getEnvelope(ctx context.Context, token string, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("core status %d: %s", resp.StatusCode, string(body))
	}

	var env Envelope[json.RawMessage]
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&env); err != nil {
		return err
	}
	if env.Code != 0 {
		return fmt.Errorf("core code %d: %s", env.Code, env.Message)
	}
	return json.Unmarshal(env.Data, out)
}
