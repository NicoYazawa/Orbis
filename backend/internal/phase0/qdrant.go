package phase0

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func qdrantURL(base, path string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("Qdrant URL must use http(s)")
	}
	suffix, query, _ := strings.Cut(path, "?")
	u.Path = strings.TrimRight(u.Path, "/") + suffix
	u.RawQuery = query
	return u.String(), nil
}
func qdrantRequest(ctx context.Context, client *http.Client, base, method, path string, body any) (*http.Response, error) {
	raw, err := qdrantURL(base, path)
	if err != nil {
		return nil, err
	}
	var b []byte
	if body != nil {
		b, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, raw, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}
func QdrantCheck(ctx context.Context, client *http.Client, base string) error {
	resp, err := qdrantRequest(ctx, client, base, http.MethodGet, "/collections", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qdrant status %d", resp.StatusCode)
	}
	return nil
}
