// Package s3 provides tests for the S3 adapter functionality.
package s3

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockS3Client implements the S3Client interface for testing purposes.
// It uses the testify/mock package to mock AWS S3 API calls.
type mockS3Client struct {
	mock.Mock
}

func (m *mockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
}

func (m *mockS3Client) GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.GetBucketLocationOutput), args.Error(1)
}

func (m *mockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

// This static assertion verifies at compile time that mockS3Client implements the S3Client interface.
var _ S3Client = (*mockS3Client)(nil)

// TestListBuckets tests the ListBuckets method of the S3 Adapter.
// It verifies that the adapter correctly processes the AWS API response
// and returns the expected list of buckets with all fields properly populated,
// including retrieving the region for each bucket.
func TestListBuckets(t *testing.T) {
	// Create mock client
	mockClient := new(mockS3Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

	// Create mock buckets
	creationDate1 := time.Now().Add(-24 * time.Hour)
	creationDate2 := time.Now().Add(-48 * time.Hour)

	// Create mock response for ListBuckets
	mockListBucketsResponse := &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				Name:         aws.String("test-bucket-1"),
				CreationDate: aws.Time(creationDate1),
			},
			{
				Name:         aws.String("test-bucket-2"),
				CreationDate: aws.Time(creationDate2),
			},
		},
	}

	// Create mock response for GetBucketLocation for bucket 1
	mockGetBucketLocationResponse1 := &s3.GetBucketLocationOutput{
		LocationConstraint: types.BucketLocationConstraintUsWest2,
	}

	// Create mock response for GetBucketLocation for bucket 2
	mockGetBucketLocationResponse2 := &s3.GetBucketLocationOutput{
		LocationConstraint: types.BucketLocationConstraintEuWest1,
	}

	// Set up expectations
	mockClient.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return(mockListBucketsResponse, nil)
	mockClient.On("GetBucketLocation", mock.Anything, &s3.GetBucketLocationInput{Bucket: aws.String("test-bucket-1")}, mock.Anything).Return(mockGetBucketLocationResponse1, nil)
	mockClient.On("GetBucketLocation", mock.Anything, &s3.GetBucketLocationInput{Bucket: aws.String("test-bucket-2")}, mock.Anything).Return(mockGetBucketLocationResponse2, nil)

	// Call the function
	ctx := context.Background()
	buckets, err := adapter.ListBuckets(ctx)

	// Assert no error
	assert.NoError(t, err)

	// Assert buckets
	assert.Len(t, buckets, 2)

	// Assert bucket 1
	assert.Equal(t, "test-bucket-1", buckets[0].Name)
	assert.Equal(t, creationDate1.Format(time.RFC3339), buckets[0].CreationDate.Format(time.RFC3339))
	assert.Equal(t, "us-west-2", buckets[0].Region)

	// Assert bucket 2
	assert.Equal(t, "test-bucket-2", buckets[1].Name)
	assert.Equal(t, creationDate2.Format(time.RFC3339), buckets[1].CreationDate.Format(time.RFC3339))
	assert.Equal(t, "eu-west-1", buckets[1].Region)

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestGetBucketRegion tests the GetBucketRegion method of the S3 Adapter.
// It verifies that the adapter correctly processes different location constraints
// and returns the appropriate region names, including the special case where
// an empty location constraint maps to "us-east-1".
func TestGetBucketRegion(t *testing.T) {
	// Create mock client
	mockClient := new(mockS3Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

	// Test cases
	testCases := []struct {
		name               string
		bucketName         string
		locationConstraint types.BucketLocationConstraint
		expectedRegion     string
	}{
		{
			name:               "US East 1 (empty constraint)",
			bucketName:         "test-bucket-us-east-1",
			locationConstraint: "",
			expectedRegion:     "us-east-1",
		},
		{
			name:               "US West 2",
			bucketName:         "test-bucket-us-west-2",
			locationConstraint: types.BucketLocationConstraintUsWest2,
			expectedRegion:     "us-west-2",
		},
		{
			name:               "EU West 1",
			bucketName:         "test-bucket-eu-west-1",
			locationConstraint: types.BucketLocationConstraintEuWest1,
			expectedRegion:     "eu-west-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock response
			mockResponse := &s3.GetBucketLocationOutput{
				LocationConstraint: tc.locationConstraint,
			}

			// Set up expectations
			mockClient.On("GetBucketLocation", mock.Anything, &s3.GetBucketLocationInput{Bucket: aws.String(tc.bucketName)}, mock.Anything).Return(mockResponse, nil).Once()

			// Call the function
			ctx := context.Background()
			region, err := adapter.GetBucketRegion(ctx, tc.bucketName)

			// Assert no error
			assert.NoError(t, err)

			// Assert region
			assert.Equal(t, tc.expectedRegion, region)
		})
	}

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestListObjects tests the ListObjects method of the S3 Adapter.
// It verifies that the adapter correctly processes the AWS API response
// and returns the expected list of objects with all fields properly populated,
// including metadata like size, ETag, storage class, and owner.
func TestListObjects(t *testing.T) {
	// Create mock client
	mockClient := new(mockS3Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

	// Create mock objects
	lastModified1 := time.Now().Add(-24 * time.Hour)
	lastModified2 := time.Now().Add(-48 * time.Hour)

	// Create mock response
	mockResponse := &s3.ListObjectsV2Output{
		Contents: []types.Object{
			{
				Key:          aws.String("test-object-1.txt"),
				Size:         aws.Int64(1024),
				LastModified: aws.Time(lastModified1),
				ETag:         aws.String("\"abc123\""),
				StorageClass: types.ObjectStorageClassStandard,
				Owner: &types.Owner{
					DisplayName: aws.String("test-owner"),
				},
			},
			{
				Key:          aws.String("test-object-2.txt"),
				Size:         aws.Int64(2048),
				LastModified: aws.Time(lastModified2),
				ETag:         aws.String("\"def456\""),
				StorageClass: types.ObjectStorageClassStandardIa,
				Owner: &types.Owner{
					DisplayName: aws.String("test-owner"),
				},
			},
		},
	}

	// Set up expectations
	mockClient.On("ListObjectsV2", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Call the function
	ctx := context.Background()
	objects, err := adapter.ListObjects(ctx, "test-bucket", "", 0)

	// Assert no error
	assert.NoError(t, err)

	// Assert objects
	assert.Len(t, objects, 2)

	// Assert object 1
	assert.Equal(t, "test-object-1.txt", objects[0].Key)
	assert.Equal(t, int64(1024), objects[0].Size)
	assert.Equal(t, lastModified1.Format(time.RFC3339), objects[0].LastModified.Format(time.RFC3339))
	assert.Equal(t, "abc123", objects[0].ETag)
	assert.Equal(t, "STANDARD", objects[0].StorageClass)
	assert.Equal(t, "test-owner", objects[0].Owner)

	// Assert object 2
	assert.Equal(t, "test-object-2.txt", objects[1].Key)
	assert.Equal(t, int64(2048), objects[1].Size)
	assert.Equal(t, lastModified2.Format(time.RFC3339), objects[1].LastModified.Format(time.RFC3339))
	assert.Equal(t, "def456", objects[1].ETag)
	assert.Equal(t, "STANDARD_IA", objects[1].StorageClass)
	assert.Equal(t, "test-owner", objects[1].Owner)

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestDeleteObject tests the DeleteObject method of the S3 Adapter.
// It verifies that the adapter correctly calls the AWS API with the
// expected parameters and handles the response.
func TestDeleteObject(t *testing.T) {
	// Create mock client
	mockClient := new(mockS3Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

	// Create mock response
	mockResponse := &s3.DeleteObjectOutput{}

	// Set up expectations
	mockClient.On("DeleteObject", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Call the function
	ctx := context.Background()
	err := adapter.DeleteObject(ctx, "test-bucket", "test-object.txt")

	// Assert no error
	assert.NoError(t, err)

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestGetObjectURL tests the GetObjectURL method of the S3 Adapter.
// It verifies that the adapter correctly formats the S3 object URL
// using the bucket name and object key.
func TestGetObjectURL(t *testing.T) {
	// Create mock client
	mockClient := new(mockS3Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

	// Call the function
	url := adapter.GetObjectURL("test-bucket", "test-object.txt")

	// Assert URL
	assert.Equal(t, "https://test-bucket.s3.amazonaws.com/test-object.txt", url)
}

// mockReadCloser implements the io.ReadCloser interface for testing purposes.
// It wraps a string reader and tracks whether Close() has been called.
// This is used to mock the response body from S3 GetObject operations.
type mockReadCloser struct {
	reader io.Reader // Underlying reader (typically a strings.Reader)
	closed bool      // Whether Close() has been called
}

// newMockReadCloser creates a new mockReadCloser with the given string content.
// This is a helper function for creating mock S3 object content in tests.
func newMockReadCloser(content string) *mockReadCloser {
	return &mockReadCloser{
		reader: strings.NewReader(content),
		closed: false,
	}
}

// Read implements the io.Reader interface by delegating to the underlying reader.
func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

// Close implements the io.Closer interface by marking the reader as closed.
// This allows tests to verify that Close() was called on the reader.
func (m *mockReadCloser) Close() error {
	m.closed = true
	return nil
}
