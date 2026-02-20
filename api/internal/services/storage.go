package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// StorageService manages Cloudflare R2 operations using the S3-compatible API.
type StorageService struct {
	client     *s3.Client
	bucket     string
	presigner  *s3.PresignClient
}

// NewStorageService initialises the R2 client.
func NewStorageService(accountID, accessKey, secretKey, bucketName, endpoint string) (*StorageService, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("storage: load config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &StorageService{
		client:    client,
		bucket:    bucketName,
		presigner: s3.NewPresignClient(client),
	}, nil
}

// PresignUpload generates a presigned PUT URL valid for 15 minutes.
// The key should follow the pattern: school-{id}/attachments/{uuid}/{filename}
func (s *StorageService) PresignUpload(ctx context.Context, key string, maxBytes int64) (string, error) {
	req, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		ContentLength: &maxBytes,
	}, func(o *s3.PresignOptions) {
		o.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("storage: presign upload: %w", err)
	}
	return req.URL, nil
}

// PresignDownload generates a presigned GET URL valid for 1 hour.
func (s *StorageService) PresignDownload(ctx context.Context, key string) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) {
		o.Expires = 1 * time.Hour
	})
	if err != nil {
		return "", fmt.Errorf("storage: presign download: %w", err)
	}
	return req.URL, nil
}

// Delete removes an object from R2.
func (s *StorageService) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("storage: delete %q: %w", key, err)
	}
	return nil
}

// ObjectKey builds a scoped R2 key for a school.
// category is one of: attachments, reports, ids, documents.
func ObjectKey(schoolID, category, filename string) string {
	return fmt.Sprintf("school-%s/%s/%s", schoolID, category, filename)
}

// ValidateMIMEType checks if the given MIME type is allowed for uploads.
func ValidateMIMEType(contentType string) error {
	allowed := map[string]bool{
		"application/pdf": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"audio/mpeg": true,
		"video/mp4":  true,
	}
	if !allowed[contentType] {
		return fmt.Errorf("storage: MIME type %q is not permitted", contentType)
	}
	return nil
}

// HeadObject retrieves metadata for an object. Used to verify upload completion.
func (s *StorageService) HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("storage: head object %q: %w", key, err)
	}
	return out, nil
}

// PutObject uploads bytes directly (used for server-generated PDFs and QR codes).
func (s *StorageService) PutObject(ctx context.Context, key string, body []byte, contentType string) error {
	r := http.NoBody
	_ = r
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        bytesReader(body),
	})
	if err != nil {
		return fmt.Errorf("storage: put object %q: %w", key, err)
	}
	return nil
}

// bytesReader wraps a byte slice in an io.Reader.
type bytesReaderWrapper struct {
	data []byte
	pos  int
}

func bytesReader(b []byte) *bytesReaderWrapper {
	return &bytesReaderWrapper{data: b}
}

func (r *bytesReaderWrapper) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
