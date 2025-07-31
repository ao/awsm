package client

import (
	"context"
	"fmt"
	"time"

	appconfig "github.com/ao/awsm/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// DefaultRetryMaxAttempts is the default number of maximum attempts for retry
const DefaultRetryMaxAttempts = 3

// DefaultRetryDelay is the default delay between retries
const DefaultRetryDelay = 100 * time.Millisecond

// Client represents an AWS client wrapper
type Client struct {
	Config aws.Config
}

// NewClient creates a new AWS client with the given options
func NewClient(ctx context.Context) (*Client, error) {
	// Get AWS profile and region from config
	profile := appconfig.GetAWSProfile()
	region := appconfig.GetAWSRegion()

	fmt.Printf("\n\nDEBUG: Creating AWS client with profile=%s, region=%s\n\n", profile, region)

	// Load AWS configuration
	cfg, err := loadConfig(ctx, profile, region)
	if err != nil {
		fmt.Printf("\n\nDEBUG: Error loading AWS config: %v\n\n", err)
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	fmt.Println("\n\nDEBUG: AWS client created successfully")

	return &Client{
		Config: cfg,
	}, nil
}

// loadConfig loads the AWS configuration with the specified profile and region
func loadConfig(ctx context.Context, profile, region string) (aws.Config, error) {
	// Create a retryer to handle retries
	retryer := retry.NewStandard(func(o *retry.StandardOptions) {
		o.MaxAttempts = DefaultRetryMaxAttempts
	})

	// Load the configuration with the specified profile and region
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithSharedConfigProfile(profile),
		awsconfig.WithRetryer(func() aws.Retryer { return retryer }),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return cfg, nil
}

// AssumeRole creates a new AWS config with assumed role credentials
func (c *Client) AssumeRole(ctx context.Context, roleARN string) (aws.Config, error) {
	// Create an STS client
	stsClient := sts.NewFromConfig(c.Config)

	// Create the credentials provider
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)

	// Create a new config with the assumed role credentials
	cfg := c.Config.Copy()
	cfg.Credentials = aws.NewCredentialsCache(provider)

	return cfg, nil
}

// GetRegion returns the region from the client config
func (c *Client) GetRegion() string {
	return c.Config.Region
}
