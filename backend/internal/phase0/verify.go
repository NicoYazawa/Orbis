package phase0

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type VerifyResult struct {
	S3Cases      int    `json:"s3_cases"`
	QdrantPoints int    `json:"qdrant_points"`
	PresignedURL string `json:"presigned_url,omitempty"`
	ObjectKey    string `json:"object_key,omitempty"`
}

func (d *Dependencies) VerifyS3(ctx context.Context, count int) (VerifyResult, error) {
	result := VerifyResult{}
	if count < 1 {
		return result, errors.New("count must be positive")
	}
	for _, size := range []int{1 << 10, 1 << 20, 10 << 20} {
		data := make([]byte, size)
		if _, err := rand.Read(data); err != nil {
			return result, err
		}
		want := sha256.Sum256(data)
		for i := 0; i < count; i++ {
			key := fmt.Sprintf("phase0-verify/%s-%d-%d", uuid.NewString(), size, i)
			_, err := d.S3.PutObject(ctx, d.Config.S3Bucket, key, bytes.NewReader(data), int64(size), minio.PutObjectOptions{ContentType: "application/octet-stream"})
			if err != nil {
				return result, fmt.Errorf("S3 put size=%d iteration=%d: %w", size, i, err)
			}
			obj, err := d.S3.GetObject(ctx, d.Config.S3Bucket, key, minio.GetObjectOptions{})
			if err != nil {
				return result, err
			}
			got, readErr := io.ReadAll(obj)
			closeErr := obj.Close()
			if readErr != nil {
				return result, readErr
			}
			if closeErr != nil {
				return result, closeErr
			}
			if sha256.Sum256(got) != want {
				return result, fmt.Errorf("S3 checksum mismatch size=%d iteration=%d", size, i)
			}
			if err := d.S3.RemoveObject(ctx, d.Config.S3Bucket, key, minio.RemoveObjectOptions{}); err != nil {
				return result, err
			}
			_, err = d.S3.StatObject(ctx, d.Config.S3Bucket, key, minio.StatObjectOptions{})
			if !isMissingObject(err) {
				return result, fmt.Errorf("S3 delete check failed size=%d iteration=%d: %v", size, i, err)
			}
			result.S3Cases++
		}
	}
	return result, nil
}

func (d *Dependencies) CreateBrowserPresign(ctx context.Context, ttl time.Duration) (VerifyResult, error) {
	result := VerifyResult{}
	if ttl < 2*time.Second {
		return result, errors.New("presign TTL must be at least 2s")
	}
	key := "phase0-browser/" + uuid.NewString() + ".txt"
	data := []byte("orbis phase0 browser presign " + key)
	if _, err := d.S3.PutObject(ctx, d.Config.S3Bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "text/plain"}); err != nil {
		return result, err
	}
	public := d.PublicS3
	if d.Config.S3PublicEndpoint != "" {
		region, err := d.S3.GetBucketLocation(ctx, d.Config.S3Bucket)
		if err != nil {
			_ = d.S3.RemoveObject(context.Background(), d.Config.S3Bucket, key, minio.RemoveObjectOptions{})
			return result, err
		}
		public, err = newS3ClientWithRegion(d.Config.S3PublicEndpoint, d.Config, region)
		if err != nil {
			_ = d.S3.RemoveObject(context.Background(), d.Config.S3Bucket, key, minio.RemoveObjectOptions{})
			return result, err
		}
	}
	signed, err := public.PresignedGetObject(ctx, d.Config.S3Bucket, key, ttl, nil)
	if err != nil {
		_ = d.S3.RemoveObject(context.Background(), d.Config.S3Bucket, key, minio.RemoveObjectOptions{})
		return result, err
	}
	result.PresignedURL = signed.String()
	result.ObjectKey = key
	return result, nil
}

func (d *Dependencies) VerifyPresign(ctx context.Context) error {
	key := "phase0-presign/" + uuid.NewString()
	data := []byte("orbis presign verification")
	if _, err := d.S3.PutObject(ctx, d.Config.S3Bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{}); err != nil {
		return err
	}
	defer d.S3.RemoveObject(context.Background(), d.Config.S3Bucket, key, minio.RemoveObjectOptions{})
	signed, err := d.S3.PresignedGetObject(ctx, d.Config.S3Bucket, key, 2*time.Second, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(signed.String())
	if err != nil {
		return err
	}
	got, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return readErr
	}
	if resp.StatusCode != 200 || !bytes.Equal(got, data) {
		return fmt.Errorf("presigned GET failed: status %d", resp.StatusCode)
	}
	tampered := *signed
	q := tampered.Query()
	q.Set("X-Amz-Signature", strings.Repeat("0", 64))
	tampered.RawQuery = q.Encode()
	resp, err = client.Get(tampered.String())
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		return fmt.Errorf("tampered signature accepted: %d", resp.StatusCode)
	}
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	resp, err = client.Get(signed.String())
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		return fmt.Errorf("expired presign accepted: %d", resp.StatusCode)
	}
	return nil
}

func (d *Dependencies) VerifyQdrant(ctx context.Context, count int) error {
	if count < 1 {
		return errors.New("point count must be positive")
	}
	collection := "orbis_phase0_verify_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	path := "/collections/" + collection
	if err := d.qdrantStatus(ctx, http.MethodPut, path, map[string]any{"vectors": map[string]any{"size": 32, "distance": "Cosine"}}); err != nil {
		return err
	}
	defer d.qdrantStatus(context.Background(), http.MethodDelete, path, nil)
	points := make([]map[string]any, 0, count)
	vectors := make([][]float64, 0, count)
	for i := 0; i < count; i++ {
		v := fixedVector(i)
		vectors = append(vectors, v)
		points = append(points, map[string]any{"id": i + 1, "vector": v})
	}
	if err := d.qdrantStatus(ctx, http.MethodPut, path+"/points?wait=true", map[string]any{"points": points}); err != nil {
		return err
	}
	for i, v := range vectors {
		resp, err := qdrantRequest(ctx, d.HTTP, d.Config.QdrantURL, http.MethodPost, path+"/points/search", map[string]any{"vector": v, "limit": 1})
		if err != nil {
			return err
		}
		var out struct {
			Result []struct {
				ID json.Number `json:"id"`
			} `json:"result"`
		}
		dec := json.NewDecoder(resp.Body)
		dec.UseNumber()
		decodeErr := dec.Decode(&out)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("qdrant search status %d", resp.StatusCode)
		}
		if decodeErr != nil {
			return decodeErr
		}
		if len(out.Result) != 1 || out.Result[0].ID.String() != fmt.Sprint(i+1) {
			return fmt.Errorf("qdrant top1 mismatch for %d", i+1)
		}
	}
	ids := make([]int, count)
	for i := range ids {
		ids[i] = i + 1
	}
	if err := d.qdrantStatus(ctx, http.MethodPost, path+"/points/delete?wait=true", map[string]any{"points": ids}); err != nil {
		return err
	}
	resp, err := qdrantRequest(ctx, d.HTTP, d.Config.QdrantURL, http.MethodPost, path+"/points/search", map[string]any{"vector": vectors[0], "limit": 1})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("qdrant deleted-point search status %d", resp.StatusCode)
	}
	var out struct {
		Result []json.RawMessage `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if len(out.Result) != 0 {
		return errors.New("deleted Qdrant points still searchable")
	}
	return nil
}

func isMissingObject(err error) bool {
	if err == nil {
		return false
	}
	return minio.ToErrorResponse(err).StatusCode == http.StatusNotFound
}

func fixedVector(i int) []float64 {
	v := make([]float64, 32)
	for j := range v {
		sum := sha256.Sum256([]byte(fmt.Sprintf("orbis-%d-%d", i, j)))
		n := new(big.Int).SetBytes(sum[:8])
		v[j] = float64(n.Uint64()>>11)/float64(uint64(1)<<53) - 0.5
	}
	norm := 0.0
	for _, x := range v {
		norm += x * x
	}
	norm = math.Sqrt(norm)
	for j := range v {
		v[j] /= norm
	}
	return v
}

func (d *Dependencies) qdrantStatus(ctx context.Context, method, path string, body any) error {
	resp, err := qdrantRequest(ctx, d.HTTP, d.Config.QdrantURL, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("qdrant %s %s status %d: %s", method, path, resp.StatusCode, b)
	}
	return nil
}
