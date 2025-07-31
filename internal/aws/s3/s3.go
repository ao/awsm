// Package s3 provides functionality for interacting with AWS S3 buckets and objects.
// It includes operations for listing buckets, listing objects, uploading, downloading,
// and deleting objects from S3 buckets.
package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ao/awsm/internal/aws/client"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client defines the interface for S3 client operations.
// This interface allows for easy mocking in tests.
type S3Client interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// Adapter represents an S3 service adapter that provides
// higher-level operations for interacting with S3 buckets and objects.
type Adapter struct {
	client S3Client // AWS S3 client implementation
}

// Bucket represents an S3 bucket with relevant information.
// This is a simplified representation of the AWS S3 bucket type
// that includes only the most commonly used fields.
type Bucket struct {
	Name         string    // Name of the bucket
	CreationDate time.Time // When the bucket was created
	Region       string    // AWS region where the bucket is located
}

// Object represents an S3 object with relevant information.
// This is a simplified representation of the AWS S3 object type
// that includes only the most commonly used fields.
type Object struct {
	Key          string    // Object key (path within the bucket)
	Size         int64     // Size of the object in bytes
	LastModified time.Time // When the object was last modified
	ETag         string    // Entity tag for the object (MD5 hash)
	StorageClass string    // Storage class of the object
	Owner        string    // Owner of the object
}

// NewAdapter creates a new S3 adapter using the AWS credentials
// from the current context configuration.
//
// The context is used for AWS client creation and configuration.
// Returns an error if the AWS client cannot be created.
func NewAdapter(ctx context.Context) (*Adapter, error) {
	// Create AWS client
	awsClient, err := client.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsClient.Config)

	return &Adapter{
		client: s3Client,
	}, nil
}

// NewAdapterWithClient creates a new S3 adapter with a provided client.
// This is particularly useful for testing with mock clients.
func NewAdapterWithClient(s3Client S3Client) *Adapter {
	return &Adapter{
		client: s3Client,
	}
}

// ListBuckets lists all S3 buckets accessible with the current credentials.
// It also attempts to determine the region for each bucket.
//
// Returns a slice of Bucket structs and an error if the operation fails.
func (a *Adapter) ListBuckets(ctx context.Context) ([]Bucket, error) {
	// Call the ListBuckets API
	output, err := a.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 buckets: %w", err)
	}

	// Extract bucket information
	buckets := make([]Bucket, 0, len(output.Buckets))
	for _, bucket := range output.Buckets {
		buckets = append(buckets, Bucket{
			Name:         aws.ToString(bucket.Name),
			CreationDate: aws.ToTime(bucket.CreationDate),
		})
	}

	// Get region for each bucket
	for i, bucket := range buckets {
		region, err := a.GetBucketRegion(ctx, bucket.Name)
		if err == nil {
			buckets[i].Region = region
		}
	}

	return buckets, nil
}

// GetBucketRegion gets the AWS region where an S3 bucket is located.
//
// Parameters:
//   - ctx: Context for the API call
//   - bucketName: The name of the S3 bucket
//
// Returns the region name as a string (e.g., "us-east-1") or an error
// if the bucket location cannot be determined.
func (a *Adapter) GetBucketRegion(ctx context.Context, bucketName string) (string, error) {
	// Call the GetBucketLocation API
	output, err := a.client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get bucket location: %w", err)
	}

	// Convert the location constraint to a region
	region := string(output.LocationConstraint)
	if region == "" {
		// Empty location constraint means us-east-1
		region = "us-east-1"
	}

	return region, nil
}

// ListObjects lists objects in an S3 bucket with optional prefix filtering.
//
// Parameters:
//   - ctx: Context for the API call
//   - bucketName: The name of the S3 bucket
//   - prefix: Optional prefix to filter objects (can be empty)
//   - maxItems: Maximum number of objects to return (0 for no limit)
//
// Returns a slice of Object structs and an error if the operation fails.
func (a *Adapter) ListObjects(ctx context.Context, bucketName, prefix string, maxItems int32) ([]Object, error) {
	// Create the input for the ListObjectsV2 API
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	// Add prefix if provided
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	// Create paginator
	paginator := s3.NewListObjectsV2Paginator(a.client, input)

	var objects []Object
	var count int32 = 0

	// Iterate through pages
	for paginator.HasMorePages() && (maxItems == 0 || count < maxItems) {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in bucket %s: %w", bucketName, err)
		}

		// Process each object
		for _, object := range output.Contents {
			// Skip if we've reached the maximum number of items
			if maxItems > 0 && count >= maxItems {
				break
			}

			// Extract object information
			obj := Object{
				Key:          aws.ToString(object.Key),
				Size:         aws.ToInt64(object.Size),
				LastModified: aws.ToTime(object.LastModified),
				ETag:         strings.Trim(aws.ToString(object.ETag), "\""),
				StorageClass: string(object.StorageClass),
			}

			// Extract owner information if available
			if object.Owner != nil && object.Owner.DisplayName != nil {
				obj.Owner = aws.ToString(object.Owner.DisplayName)
			}

			objects = append(objects, obj)
			count++
		}
	}

	return objects, nil
}

// UploadObject uploads a local file to an S3 bucket.
//
// Parameters:
//   - ctx: Context for the API call
//   - bucketName: The name of the S3 bucket
//   - key: The key (path) to store the object under in the bucket
//   - filePath: The local file path to upload
//
// Returns an error if the file cannot be opened or the upload fails.
func (a *Adapter) UploadObject(ctx context.Context, bucketName, key, filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Create the input for the PutObject API
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   file,
	}

	// Call the PutObject API
	_, err = a.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload object to bucket %s: %w", bucketName, err)
	}

	return nil
}

// DownloadObject downloads an object from an S3 bucket to a local file.
// It will create any necessary directories in the file path if they don't exist.
//
// Parameters:
//   - ctx: Context for the API call
//   - bucketName: The name of the S3 bucket
//   - key: The key (path) of the object in the bucket
//   - filePath: The local file path to save the object to
//
// Returns an error if the directories cannot be created, the file cannot be created,
// or the download fails.
func (a *Adapter) DownloadObject(ctx context.Context, bucketName, key, filePath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Create the input for the GetObject API
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	// Call the GetObject API
	output, err := a.client.GetObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to download object from bucket %s: %w", bucketName, err)
	}
	defer output.Body.Close()

	// Copy the object data to the file
	_, err = io.Copy(file, output.Body)
	if err != nil {
		return fmt.Errorf("failed to write object data to file: %w", err)
	}

	return nil
}

// DeleteObject deletes an object from an S3 bucket.
//
// Parameters:
//   - ctx: Context for the API call
//   - bucketName: The name of the S3 bucket
//   - key: The key (path) of the object to delete
//
// Returns an error if the deletion fails.
func (a *Adapter) DeleteObject(ctx context.Context, bucketName, key string) error {
	// Create the input for the DeleteObject API
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	// Call the DeleteObject API
	_, err := a.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object from bucket %s: %w", bucketName, err)
	}

	return nil
}

// GetObjectURL gets the public URL of an S3 object.
// Note that this does not check if the object exists or if it's publicly accessible.
//
// Parameters:
//   - bucketName: The name of the S3 bucket
//   - key: The key (path) of the object in the bucket
//
// Returns the URL as a string.
func (a *Adapter) GetObjectURL(bucketName, key string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, key)
}
