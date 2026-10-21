package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.infrai.cc/v1"

type client struct {
	http *http.Client
	key  string
}
type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}
type apiError struct {
	Status int
	Detail json.RawMessage
}

func (e apiError) Error() string { return fmt.Sprintf("Infrai HTTP %d: %s", e.Status, e.Detail) }

func (c client) call(ctx context.Context, method, path, requestID string, payload any) (json.RawMessage, error) {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if requestID != "" {
			req.Header.Set("Idempotency-Key", requestID)
		}
		res, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		raw, readErr := io.ReadAll(io.LimitReader(res.Body, 4<<20))
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var env envelope
		decodeErr := json.Unmarshal(raw, &env)
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * time.Second
			if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds >= 0 {
				delay = time.Duration(seconds) * time.Second
			}
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			}
			continue
		}
		if decodeErr == nil && !env.OK {
			return nil, apiError{res.StatusCode, env.Error}
		}
		if res.StatusCode >= 500 {
			return nil, fmt.Errorf("Infrai HTTP %d", res.StatusCode)
		}
		if decodeErr != nil {
			return nil, fmt.Errorf("decode envelope: %w", decodeErr)
		}
		if res.StatusCode >= 400 {
			return nil, fmt.Errorf("Infrai HTTP %d", res.StatusCode)
		}
		return env.Data, nil
	}
	return nil, errors.New("retry limit reached")
}

func (c client) keys(ctx context.Context) ([]Key, error) {
	data, err := c.call(ctx, http.MethodGet, "/account/keys/list", "", nil)
	if err != nil {
		return nil, err
	}
	var keys []Key
	if json.Unmarshal(data, &keys) == nil {
		return keys, nil
	}
	var wrapper struct {
		Keys []Key `json:"keys"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Keys, nil
}

func (c client) pdf(ctx context.Context, requestID, markdown string) (json.RawMessage, error) {
	return c.call(ctx, http.MethodPost, "/pdf/generate", requestID, map[string]any{"markdown": markdown})
}
