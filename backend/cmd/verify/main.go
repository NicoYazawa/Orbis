package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/NicoYazawa/Orbis/internal/phase0"
	"github.com/minio/minio-go/v7"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	healthURL := flag.String("health-url", "", "GET liveness URL with a 2 second timeout")
	count := flag.Int("count", 100, "S3 CRUD iterations per size")
	pointCount := flag.Int("qdrant-count", 1000, "Qdrant vector count")
	presignOnly := flag.Bool("presign-only", false, "create an object and emit a public presigned GET URL as JSON")
	deleteKey := flag.String("delete-key", "", "delete a browser verification object key")
	flag.Parse()
	if *healthURL != "" {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(*healthURL)
		if err != nil {
			return fmt.Errorf("healthcheck failed: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("healthcheck status %d", resp.StatusCode)
		}
		return nil
	}
	c, err := phase0.LoadConfig(os.Getenv)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	d, err := phase0.Open(ctx, c)
	if err != nil {
		return fmt.Errorf("dependency connection failed")
	}
	defer d.Close()
	if *deleteKey != "" {
		return d.S3.RemoveObject(ctx, c.S3Bucket, *deleteKey, minio.RemoveObjectOptions{})
	}
	if *presignOnly {
		result, err := d.CreateBrowserPresign(ctx, 15*time.Minute)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	result, err := d.VerifyS3(ctx, *count)
	if err != nil {
		return err
	}
	if err := d.VerifyPresign(ctx); err != nil {
		return fmt.Errorf("S3 presign: %w", err)
	}
	if c.QdrantURL != "" {
		if err := d.VerifyQdrant(ctx, *pointCount); err != nil {
			return err
		}
		result.QdrantPoints = *pointCount
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
