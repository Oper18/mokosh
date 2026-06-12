package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// S3Backend implements Backend for AWS S3 and S3-compatible object stores
// (Hetzner Object Storage, MinIO, Cloudflare R2, DigitalOcean Spaces, …).
type S3Backend struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

// NewS3 initialises an S3-backed storage backend.
// endpoint may be left empty for native AWS S3; for other providers supply
// the full base URL, e.g. "https://fsn1.your-objectstorage.com".
// Path-style addressing is always enabled so the same config works for
// AWS and any S3-compatible provider.
func NewS3(endpoint, region, accessKeyID, secretAccessKey, bucket string) (*S3Backend, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	clientOpts := []func(*s3.Options){
		func(o *s3.Options) {
			// Path-style URLs work with all S3-compatible providers.
			o.UsePathStyle = true
		},
	}

	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	client := s3.NewFromConfig(cfg, clientOpts...)

	return &S3Backend{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  bucket,
	}, nil
}

// Put uploads the content from r to the S3 object at key.
// Pass size = -1 to use chunked / streaming transfer encoding.
func (s *S3Backend) Put(ctx context.Context, key string, r io.Reader, size int64) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	}

	if size >= 0 {
		input.ContentLength = aws.Int64(size)
	}

	_, err := s.client.PutObject(ctx, input)
	return err
}

// Exists checks whether an object with the given key exists in the bucket.
//
// It returns (false, nil) only when the store positively reports the object as
// absent (HTTP 404). Any other failure (network error, auth, throttling, an
// unreachable endpoint) is returned to the caller so that a temporarily
// unreachable store is never mistaken for a deleted object — callers that purge
// missing files rely on this distinction to avoid destroying remote-backed data.
func (s *S3Backend) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}

	// A 404 response means the object genuinely does not exist.
	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusNotFound {
		return false, nil
	}

	// Any other error is treated as "unknown", not "absent".
	return false, err
}

// PresignedGetURL returns a pre-authenticated GET URL for the given key.
// The URL expires after expiry.
func (s *S3Backend) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
