// Package integration provides integration tests for the awsm CLI commands.
// These tests verify that the CLI commands correctly interact with AWS services
// using mock implementations of the AWS clients.
package integration

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/ao/awsm/internal/aws/ec2"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockEC2Client implements the ec2.EC2Client interface for testing purposes.
// It uses the testify/mock package to mock AWS EC2 API calls, allowing
// tests to run without making actual AWS API requests.
type mockEC2Client struct {
	mock.Mock
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *awsec2.DescribeInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsec2.DescribeInstancesOutput), args.Error(1)
}

func (m *mockEC2Client) StartInstances(ctx context.Context, params *awsec2.StartInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.StartInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsec2.StartInstancesOutput), args.Error(1)
}

func (m *mockEC2Client) StopInstances(ctx context.Context, params *awsec2.StopInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.StopInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsec2.StopInstancesOutput), args.Error(1)
}

// This static assertion verifies at compile time that mockEC2Client implements the ec2.EC2Client interface.
var _ ec2.EC2Client = (*mockEC2Client)(nil)

// executeCommand is a helper function that executes a Cobra command with the given arguments
// and captures its output. This is useful for testing CLI commands without actually
// executing the real command implementation.
//
// Parameters:
//   - root: The root Cobra command to execute
//   - args: Command line arguments to pass to the command
//
// Returns the command output as a string and any error that occurred during execution.
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

// createMockInstance is a helper function that creates a mock EC2 instance with the specified parameters.
// This function simplifies the creation of test data for EC2 instance tests.
//
// Parameters:
//   - id: The EC2 instance ID
//   - name: The Name tag value
//   - instanceType: The EC2 instance type (e.g., t2.micro)
//   - state: The instance state (e.g., running, stopped)
//   - publicIP: The public IP address (can be empty)
//   - privateIP: The private IP address (can be empty)
//   - az: The availability zone
//   - vpcID: The VPC ID
//   - subnetID: The subnet ID
//   - tags: Additional tags to add to the instance
//
// Returns a types.Instance that can be used in test mocks.
func createMockInstance(id, name, instanceType, state, publicIP, privateIP, az, vpcID, subnetID string, tags map[string]string) types.Instance {
	// Create tags
	var instanceTags []types.Tag
	for k, v := range tags {
		key := k
		value := v
		instanceTags = append(instanceTags, types.Tag{
			Key:   &key,
			Value: &value,
		})
	}

	// Add name tag if provided
	if name != "" {
		nameKey := "Name"
		instanceTags = append(instanceTags, types.Tag{
			Key:   &nameKey,
			Value: &name,
		})
	}

	// Create security groups
	securityGroups := []types.GroupIdentifier{
		{
			GroupId:   aws.String("sg-12345"),
			GroupName: aws.String("default"),
		},
	}

	// Create instance
	instance := types.Instance{
		InstanceId:   aws.String(id),
		InstanceType: types.InstanceType(instanceType),
		State: &types.InstanceState{
			Name: types.InstanceStateName(state),
		},
		Placement: &types.Placement{
			AvailabilityZone: aws.String(az),
		},
		VpcId:          aws.String(vpcID),
		SubnetId:       aws.String(subnetID),
		Tags:           instanceTags,
		SecurityGroups: securityGroups,
	}

	// Add IPs if provided
	if publicIP != "" {
		instance.PublicIpAddress = aws.String(publicIP)
	}
	if privateIP != "" {
		instance.PrivateIpAddress = aws.String(privateIP)
	}

	return instance
}

// TestEC2ListCommand tests the EC2 list command functionality.
// It verifies that the command correctly retrieves and displays EC2 instances
// using a mock EC2 client that returns predefined instance data.
// The test checks that the output contains the expected instance information.
func TestEC2ListCommand(t *testing.T) {
	// Skip this test if running in CI
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Create a mock EC2 client
	mockClient := new(mockEC2Client)

	// Create mock instances
	instance1 := createMockInstance(
		"i-12345",
		"test-instance-1",
		"t2.micro",
		"running",
		"1.2.3.4",
		"10.0.0.1",
		"us-east-1a",
		"vpc-12345",
		"subnet-12345",
		map[string]string{"Environment": "test"},
	)

	instance2 := createMockInstance(
		"i-67890",
		"test-instance-2",
		"t2.small",
		"stopped",
		"",
		"10.0.0.2",
		"us-east-1b",
		"vpc-12345",
		"subnet-67890",
		map[string]string{"Environment": "prod"},
	)

	// Create mock response
	mockResponse := &awsec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{instance1},
			},
			{
				Instances: []types.Instance{instance2},
			},
		},
	}

	// Set up expectations
	mockClient.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Create a new EC2 adapter with the mock client
	adapter := ec2.NewAdapterWithClient(mockClient)

	// Create a new EC2 command
	ec2Cmd := &cobra.Command{
		Use:   "ec2",
		Short: "EC2 instance management",
	}

	// Add the list subcommand
	ec2Cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List EC2 instances",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			instances, err := adapter.ListInstances(ctx, nil, 0)
			if err != nil {
				cmd.PrintErrf("Error: %s", err)
				return
			}
			for _, instance := range instances {
				cmd.Printf("ID: %s, Name: %s, State: %s\n", instance.ID, instance.Name, instance.State)
			}
		},
	})

	// Execute the command
	output, err := executeCommand(ec2Cmd, "list")

	// Assert no error
	assert.NoError(t, err)

	// Assert output contains instance information
	assert.Contains(t, output, "ID: i-12345, Name: test-instance-1, State: running")
	assert.Contains(t, output, "ID: i-67890, Name: test-instance-2, State: stopped")

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestEC2StartCommand tests the EC2 start command functionality.
// It verifies that the command correctly starts an EC2 instance
// using a mock EC2 client that simulates a successful start operation.
// The test checks that the output contains the expected success message.
func TestEC2StartCommand(t *testing.T) {
	// Skip this test if running in CI
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Create a mock EC2 client
	mockClient := new(mockEC2Client)

	// Create mock response
	mockResponse := &awsec2.StartInstancesOutput{
		StartingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String("i-12345"),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameStopped,
				},
			},
		},
	}

	// Set up expectations
	mockClient.On("StartInstances", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Create a new EC2 adapter with the mock client
	adapter := ec2.NewAdapterWithClient(mockClient)

	// Create a new EC2 command
	ec2Cmd := &cobra.Command{
		Use:   "ec2",
		Short: "EC2 instance management",
	}

	// Add the start subcommand
	ec2Cmd.AddCommand(&cobra.Command{
		Use:   "start [instance-id]",
		Short: "Start an EC2 instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			instanceID := args[0]
			err := adapter.StartInstance(ctx, instanceID)
			if err != nil {
				cmd.PrintErrf("Error: %s", err)
				return
			}
			cmd.Printf("Successfully started EC2 instance %s\n", instanceID)
		},
	})

	// Execute the command
	output, err := executeCommand(ec2Cmd, "start", "i-12345")

	// Assert no error
	assert.NoError(t, err)

	// Assert output contains success message
	assert.Contains(t, output, "Successfully started EC2 instance i-12345")

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestEC2StopCommand tests the EC2 stop command functionality.
// It verifies that the command correctly stops an EC2 instance
// using a mock EC2 client that simulates a successful stop operation.
// The test checks that the output contains the expected success message.
func TestEC2StopCommand(t *testing.T) {
	// Skip this test if running in CI
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Create a mock EC2 client
	mockClient := new(mockEC2Client)

	// Create mock response
	mockResponse := &awsec2.StopInstancesOutput{
		StoppingInstances: []types.InstanceStateChange{
			{
				InstanceId: aws.String("i-12345"),
				CurrentState: &types.InstanceState{
					Name: types.InstanceStateNameStopping,
				},
				PreviousState: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
		},
	}

	// Set up expectations
	mockClient.On("StopInstances", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Create a new EC2 adapter with the mock client
	adapter := ec2.NewAdapterWithClient(mockClient)

	// Create a new EC2 command
	ec2Cmd := &cobra.Command{
		Use:   "ec2",
		Short: "EC2 instance management",
	}

	// Add the stop subcommand
	ec2Cmd.AddCommand(&cobra.Command{
		Use:   "stop [instance-id]",
		Short: "Stop an EC2 instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			instanceID := args[0]
			err := adapter.StopInstance(ctx, instanceID)
			if err != nil {
				cmd.PrintErrf("Error: %s", err)
				return
			}
			cmd.Printf("Successfully stopped EC2 instance %s\n", instanceID)
		},
	})

	// Execute the command
	output, err := executeCommand(ec2Cmd, "stop", "i-12345")

	// Assert no error
	assert.NoError(t, err)

	// Assert output contains success message
	assert.Contains(t, output, "Successfully stopped EC2 instance i-12345")

	// Verify expectations
	mockClient.AssertExpectations(t)
}
