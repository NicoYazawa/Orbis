package phase0

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string `gorm:"size:100;not null"`
	Role         string `gorm:"size:32;not null;default:user"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DBAdminStore struct{ DB *gorm.DB }

func (s DBAdminStore) CreateAdminIfAbsent(ctx context.Context, username, hash string) error {
	u := User{Username: username, PasswordHash: hash, Role: "admin"}
	return s.DB.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "username"}}, DoNothing: true}).Create(&u).Error
}

type Dependencies struct {
	DB       *gorm.DB
	Redis    *redis.Client
	S3       *minio.Client
	PublicS3 *minio.Client
	HTTP     *http.Client
	Config   Config
}

func Open(ctx context.Context, c Config) (*Dependencies, error) {
	db, err := gorm.Open(postgres.Open(c.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	rdb := redis.NewClient(&redis.Options{Addr: c.RedisAddr})
	s3, err := newS3Client(c.S3Endpoint, c)
	if err != nil {
		return nil, err
	}
	var public *minio.Client
	if c.S3PublicEndpoint != "" {
		public, err = newS3Client(c.S3PublicEndpoint, c)
		if err != nil {
			return nil, fmt.Errorf("public S3: %w", err)
		}
	} else {
		public = s3
	}
	d := &Dependencies{DB: db, Redis: rdb, S3: s3, PublicS3: public, HTTP: &http.Client{Timeout: 3 * time.Second}, Config: c}
	return d, nil
}

func newS3Client(raw string, c Config) (*minio.Client, error) {
	return newS3ClientWithRegion(raw, c, "")
}

func newS3ClientWithRegion(raw string, c Config, region string) (*minio.Client, error) {
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "" {
		return nil, errors.New("S3 endpoint must be an HTTP(S) origin without a path")
	}
	return minio.New(u.Host, &minio.Options{Creds: credentials.NewStaticV4(c.S3AccessKey, c.S3SecretKey, ""), Secure: u.Scheme == "https", Region: region})
}

func (d *Dependencies) Close() {
	if d.Redis != nil {
		_ = d.Redis.Close()
	}
	if d.DB != nil {
		if sqlDB, err := d.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}

func (d *Dependencies) Checks() map[string]Probe {
	checks := map[string]Probe{
		"postgres": ProbeFunc(func(ctx context.Context) error {
			sqlDB, err := d.DB.DB()
			if err != nil {
				return err
			}
			return sqlDB.PingContext(ctx)
		}),
		"redis": ProbeFunc(func(ctx context.Context) error { return d.Redis.Ping(ctx).Err() }),
		"s3": ProbeFunc(func(ctx context.Context) error {
			exists, err := d.S3.BucketExists(ctx, d.Config.S3Bucket)
			if err != nil {
				return err
			}
			if !exists {
				return errors.New("bucket absent")
			}
			return nil
		}),
	}
	if d.Config.RequireRAG {
		checks["qdrant"] = ProbeFunc(func(ctx context.Context) error { return QdrantCheck(ctx, d.HTTP, d.Config.QdrantURL) })
	}
	return checks
}

func (d *Dependencies) Initialize(ctx context.Context) error {
	if err := ValidateAdminCredentials(d.Config.AdminUsername, d.Config.AdminPassword); err != nil {
		return err
	}
	if err := d.Migrate(ctx); err != nil {
		return err
	}
	if err := EnsureAdmin(ctx, DBAdminStore{d.DB}, d.Config.AdminUsername, d.Config.AdminPassword); err != nil {
		return err
	}
	exists, err := d.S3.BucketExists(ctx, d.Config.S3Bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := d.S3.MakeBucket(ctx, d.Config.S3Bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	return nil
}
