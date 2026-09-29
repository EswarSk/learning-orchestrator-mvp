package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	Base string
	HTTP *http.Client
}
type RemoteError struct {
	Status int
	Body   string
}

func (e RemoteError) Error() string { return fmt.Sprintf("remote HTTP %d: %s", e.Status, e.Body) }

func NewClient(env, fallback string) Client {
	return Client{Base: strings.TrimRight(Env(env, fallback), "/"), HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func (c Client) Do(ctx context.Context, user, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.Base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Token", os.Getenv("INTERNAL_TOKEN"))
	req.Header.Set("X-User-ID", user)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return RemoteError{Status: res.StatusCode, Body: string(data)}
	}
	if output != nil {
		return json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(output)
	}
	return nil
}
