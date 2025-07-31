// Package ec2 provides functionality for interacting with AWS EC2 instances.
// It includes operations for listing, describing, starting, and stopping EC2 instances.
package ec2

import (
	"context"
	"fmt"

	"github.com/ao/awsm/internal/aws/client"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Client defines the interface for EC2 client operations.
// This interface allows for easy mocking in tests.
type EC2Client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
}

// Adapter represents an EC2 service adapter that provides
// higher-level operations for interacting with EC2 instances.
type Adapter struct {
	client EC2Client // AWS EC2 client implementation
}

// Instance represents an EC2 instance with relevant information.
// This is a simplified representation of the AWS EC2 instance type
// that includes only the most commonly used fields.
type Instance struct {
	ID          string            // EC2 instance ID (i-xxxxxxxx)
	Name        string            // Name tag value if available
	Type        string            // Instance type (e.g., t2.micro)
	State       string            // Current state (running, stopped, etc.)
	PublicIP    string            // Public IP address if available
	PrivateIP   string            // Private IP address
	LaunchTime  string            // When the instance was launched (formatted)
	AZ          string            // Availability Zone
	VpcID       string            // VPC ID
	SubnetID    string            // Subnet ID
	Tags        map[string]string // All instance tags
	SecurityIDs []string          // Security group IDs
}

// NewAdapter creates a new EC2 adapter using the AWS credentials
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

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(awsClient.Config)

	return &Adapter{
		client: ec2Client,
	}, nil
}

// NewAdapterWithClient creates a new EC2 adapter with a provided client.
// This is particularly useful for testing with mock clients.
func NewAdapterWithClient(ec2Client EC2Client) *Adapter {
	return &Adapter{
		client: ec2Client,
	}
}

// ListInstances lists EC2 instances with optional filtering.
//
// Parameters:
//   - ctx: Context for the API call
//   - filters: Optional EC2 filters to apply (can be nil or empty)
//   - maxItems: Maximum number of instances to return (0 for no limit)
//
// Returns a slice of Instance structs and an error if the operation fails.
func (a *Adapter) ListInstances(ctx context.Context, filters []types.Filter, maxItems int32) ([]Instance, error) {
	// Create the input for the DescribeInstances API
	input := &ec2.DescribeInstancesInput{}

	// Add filters if provided
	if len(filters) > 0 {
		input.Filters = filters
	}

	// Call the DescribeInstances API
	paginator := ec2.NewDescribeInstancesPaginator(a.client, input)

	var instances []Instance
	var count int32 = 0

	// Iterate through pages
	for paginator.HasMorePages() && (maxItems == 0 || count < maxItems) {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list EC2 instances: %w", err)
		}

		// Process each reservation
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				// Skip if we've reached the maximum number of items
				if maxItems > 0 && count >= maxItems {
					break
				}

				// Extract instance information
				inst := extractInstanceInfo(instance)
				instances = append(instances, inst)
				count++
			}
		}
	}

	return instances, nil
}

// DescribeInstance gets detailed information about a specific EC2 instance.
//
// Parameters:
//   - ctx: Context for the API call
//   - instanceID: The ID of the EC2 instance to describe
//
// Returns a pointer to an Instance struct with the instance details
// or an error if the instance cannot be found or described.
func (a *Adapter) DescribeInstance(ctx context.Context, instanceID string) (*Instance, error) {
	// Create the input for the DescribeInstances API
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Call the DescribeInstances API
	output, err := a.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instance %s: %w", instanceID, err)
	}

	// Check if the instance was found
	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("EC2 instance %s not found", instanceID)
	}

	// Extract instance information
	instance := extractInstanceInfo(output.Reservations[0].Instances[0])

	return &instance, nil
}

// StartInstance starts an EC2 instance.
//
// Parameters:
//   - ctx: Context for the API call
//   - instanceID: The ID of the EC2 instance to start
//
// Returns an error if the instance cannot be started.
func (a *Adapter) StartInstance(ctx context.Context, instanceID string) error {
	// Create the input for the StartInstances API
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Call the StartInstances API
	_, err := a.client.StartInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start EC2 instance %s: %w", instanceID, err)
	}

	return nil
}

// StopInstance stops an EC2 instance.
//
// Parameters:
//   - ctx: Context for the API call
//   - instanceID: The ID of the EC2 instance to stop
//
// Returns an error if the instance cannot be stopped.
func (a *Adapter) StopInstance(ctx context.Context, instanceID string) error {
	// Create the input for the StopInstances API
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Call the StopInstances API
	_, err := a.client.StopInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to stop EC2 instance %s: %w", instanceID, err)
	}

	return nil
}

// extractInstanceInfo extracts relevant information from an EC2 instance
// and converts it to our simplified Instance struct.
//
// This is an internal helper function used by ListInstances and DescribeInstance.
func extractInstanceInfo(instance types.Instance) Instance {
	// Initialize the instance
	inst := Instance{
		ID:          aws.ToString(instance.InstanceId),
		Type:        string(instance.InstanceType),
		State:       string(instance.State.Name),
		AZ:          aws.ToString(instance.Placement.AvailabilityZone),
		VpcID:       aws.ToString(instance.VpcId),
		SubnetID:    aws.ToString(instance.SubnetId),
		Tags:        make(map[string]string),
		SecurityIDs: make([]string, 0),
	}

	// Extract IP addresses if available
	if instance.PublicIpAddress != nil {
		inst.PublicIP = *instance.PublicIpAddress
	}
	if instance.PrivateIpAddress != nil {
		inst.PrivateIP = *instance.PrivateIpAddress
	}

	// Extract launch time if available
	if instance.LaunchTime != nil {
		inst.LaunchTime = instance.LaunchTime.Format("2006-01-02 15:04:05")
	}

	// Extract tags
	for _, tag := range instance.Tags {
		if tag.Key != nil && tag.Value != nil {
			inst.Tags[*tag.Key] = *tag.Value
			// Set the Name tag as the instance name
			if *tag.Key == "Name" {
				inst.Name = *tag.Value
			}
		}
	}

	// Extract security group IDs
	for _, sg := range instance.SecurityGroups {
		if sg.GroupId != nil {
			inst.SecurityIDs = append(inst.SecurityIDs, *sg.GroupId)
		}
	}

	return inst
}

// CreateFilter creates an EC2 filter for use with the ListInstances function.
//
// Parameters:
//   - name: The name of the filter (e.g., "instance-state-name")
//   - values: One or more values for the filter (e.g., "running", "stopped")
//
// Returns an EC2 filter that can be used in API calls.
func CreateFilter(name string, values ...string) types.Filter {
	return types.Filter{
		Name:   aws.String(name),
		Values: values,
	}
}
