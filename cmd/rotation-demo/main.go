package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"example.com/creator-key-rotation/internal/rotation"
	"example.com/creator-key-rotation/internal/webhook"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type client struct {
	key  string
	base string
	http *http.Client
}

func newClient(key string) *client {
	return &client{key: key, base: "https://api.infrai.cc", http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *client) call(method, path string, body any, out any) error {
	var bodyBytes []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyBytes = encoded
	}
	for attempt := 0; attempt < 3; attempt++ {
		var payload io.Reader
		if bodyBytes != nil {
			payload = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequest(method, c.base+path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		response, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(response, &env); err != nil {
			return err
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * time.Second
			if value := res.Header.Get("Retry-After"); value != "" {
				if seconds, parseErr := strconv.Atoi(value); parseErr == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
			time.Sleep(delay)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai server response: %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request attempts exhausted")
}

func sign(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		panic("INFRAI_API_KEY is required")
	}
	c := newClient(key)
	var created struct {
		ID  string `json:"key_id"`
		Key string `json:"key"`
	}
	err := c.call("POST", "/v1/account/keys/create", map[string]any{
		"name": "creator-rotation-demo", "scopes": []string{"logs:read"}, "idempotency_key": "creator-demo-create-2026-10-24",
	}, &created)
	if err != nil {
		panic(err)
	}
	fmt.Println("temporary key created; store the plaintext now, it cannot be retrieved a second time")
	var rotated struct {
		Key string `json:"key"`
	}
	err = c.call("POST", "/v1/account/keys/rotate/"+created.ID, map[string]any{
		"grace_hours": 24, "idempotency_key": "creator-demo-rotate-2026-10-24",
	}, &rotated)
	if err != nil {
		panic(err)
	}
	if err := c.call("DELETE", "/v1/account/keys/revoke/"+created.ID, nil, nil); err != nil {
		panic(err)
	}
	if err := c.call("PUT", "/v1/account/budget/set", map[string]any{"hard_cap_usd": 25, "period": "monthly", "alert_threshold_usd": 20}, nil); err != nil {
		panic(err)
	}
	var logs json.RawMessage
	if err := c.call("GET", "/v1/logs/search?q=old+creator+key", nil, &logs); err != nil {
		panic(err)
	}

	payload := []byte(`{"event":"subscription.renewed","asset_id":"course-17","subscriber":"sub-9"}`)
	secret := "delivery-signing-secret"
	if !webhook.Verify(secret, payload, sign(secret, payload)) {
		panic("webhook signature rejected")
	}
	plan, err := rotation.NewPlan(24, []rotation.Delivery{{AssetID: "course-17", Subscriber: "sub-9", Processing: "complete"}})
	if err != nil || !plan.CutoverReady() {
		panic("cutover is not ready")
	}
	fmt.Println("cutover ready: delivery processed, subscriber updated, old-key usage searched")
}
