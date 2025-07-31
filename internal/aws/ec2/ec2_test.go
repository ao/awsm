// Package ec2 provides tests for the EC2 adapter functionality.
package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockEC2Client implements the EC2Client interface for testing purposes.
// It uses the testify/mock package to mock AWS EC2 API calls.
type mockEC2Client struct {
	mock.Mock
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
}

func (m *mockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ec2.StartInstancesOutput), args.Error(1)
}

func (m *mockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ec2.StopInstancesOutput), args.Error(1)
}

// This static assertion verifies at compile time that mockEC2Client implements the EC2Client interface.
var _ EC2Client = (*mockEC2Client)(nil)

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

	// Create launch time
	launchTime := time.Now().Add(-24 * time.Hour)

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
		LaunchTime:     &launchTime,
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

// TestListInstances tests the ListInstances method of the EC2 Adapter.
// It verifies that the adapter correctly processes the AWS API response
// and returns the expected list of instances with all fields properly populated.
func TestListInstances(t *testing.T) {
	// Create mock client
	mockClient := new(mockEC2Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

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
	mockResponse := &ec2.DescribeInstancesOutput{
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

	// Call the function
	ctx := context.Background()
	instances, err := adapter.ListInstances(ctx, nil, 0)

	// Assert no error
	assert.NoError(t, err)

	// Assert instances
	assert.Len(t, instances, 2)

	// Assert instance 1
	assert.Equal(t, "i-12345", instances[0].ID)
	assert.Equal(t, "test-instance-1", instances[0].Name)
	assert.Equal(t, "t2.micro", instances[0].Type)
	assert.Equal(t, "running", instances[0].State)
	assert.Equal(t, "1.2.3.4", instances[0].PublicIP)
	assert.Equal(t, "10.0.0.1", instances[0].PrivateIP)
	assert.Equal(t, "us-east-1a", instances[0].AZ)
	assert.Equal(t, "vpc-12345", instances[0].VpcID)
	assert.Equal(t, "subnet-12345", instances[0].SubnetID)
	assert.Equal(t, "test", instances[0].Tags["Environment"])

	// Assert instance 2
	assert.Equal(t, "i-67890", instances[1].ID)
	assert.Equal(t, "test-instance-2", instances[1].Name)
	assert.Equal(t, "t2.small", instances[1].Type)
	assert.Equal(t, "stopped", instances[1].State)
	assert.Equal(t, "", instances[1].PublicIP)
	assert.Equal(t, "10.0.0.2", instances[1].PrivateIP)
	assert.Equal(t, "us-east-1b", instances[1].AZ)
	assert.Equal(t, "vpc-12345", instances[1].VpcID)
	assert.Equal(t, "subnet-67890", instances[1].SubnetID)
	assert.Equal(t, "prod", instances[1].Tags["Environment"])

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestDescribeInstance tests the DescribeInstance method of the EC2 Adapter.
// It verifies that the adapter correctly processes the AWS API response
// and returns the expected instance details.
func TestDescribeInstance(t *testing.T) {
	// Create mock client
	mockClient := new(mockEC2Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

	// Create mock instance
	instance := createMockInstance(
		"i-12345",
		"test-instance",
		"t2.micro",
		"running",
		"1.2.3.4",
		"10.0.0.1",
		"us-east-1a",
		"vpc-12345",
		"subnet-12345",
		map[string]string{"Environment": "test"},
	)

	// Create mock response
	mockResponse := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{instance},
			},
		},
	}

	// Set up expectations
	mockClient.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Call the function
	ctx := context.Background()
	result, err := adapter.DescribeInstance(ctx, "i-12345")

	// Assert no error
	assert.NoError(t, err)

	// Assert instance
	assert.Equal(t, "i-12345", result.ID)
	assert.Equal(t, "test-instance", result.Name)
	assert.Equal(t, "t2.micro", result.Type)
	assert.Equal(t, "running", result.State)
	assert.Equal(t, "1.2.3.4", result.PublicIP)
	assert.Equal(t, "10.0.0.1", result.PrivateIP)
	assert.Equal(t, "us-east-1a", result.AZ)
	assert.Equal(t, "vpc-12345", result.VpcID)
	assert.Equal(t, "subnet-12345", result.SubnetID)
	assert.Equal(t, "test", result.Tags["Environment"])

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestStartInstance tests the StartInstance method of the EC2 Adapter.
// It verifies that the adapter correctly calls the AWS API with the
// expected parameters and handles the response.
func TestStartInstance(t *testing.T) {
	// Create mock client
	mockClient := new(mockEC2Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

	// Create mock response
	mockResponse := &ec2.StartInstancesOutput{
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

	// Call the function
	ctx := context.Background()
	err := adapter.StartInstance(ctx, "i-12345")

	// Assert no error
	assert.NoError(t, err)

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestStopInstance tests the StopInstance method of the EC2 Adapter.
// It verifies that the adapter correctly calls the AWS API with the
// expected parameters and handles the response.
func TestStopInstance(t *testing.T) {
	// Create mock client
	mockClient := new(mockEC2Client)

	// Create adapter with mock client
	adapter := NewAdapterWithClient(mockClient)

	// Create mock response
	mockResponse := &ec2.StopInstancesOutput{
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

	// Call the function
	ctx := context.Background()
	err := adapter.StopInstance(ctx, "i-12345")

	// Assert no error
	assert.NoError(t, err)

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestCreateFilter tests the CreateFilter function.
// It verifies that the function correctly creates an EC2 filter
// with the specified name and values.
func TestCreateFilter(t *testing.T) {
	// Create a filter
	filter := CreateFilter("tag:Name", "test-instance")

	// Assert filter
	assert.Equal(t, "tag:Name", *filter.Name)
	assert.Equal(t, []string{"test-instance"}, filter.Values)
}

// TestExtractInstanceInfo tests the extractInstanceInfo function.
// It verifies that the function correctly extracts information from
// an AWS EC2 instance and converts it to our simplified Instance struct.
func TestExtractInstanceInfo(t *testing.T) {
	// Create mock instance
	instance := createMockInstance(
		"i-12345",
		"test-instance",
		"t2.micro",
		"running",
		"1.2.3.4",
		"10.0.0.1",
		"us-east-1a",
		"vpc-12345",
		"subnet-12345",
		map[string]string{"Environment": "test"},
	)

	// Call the function
	result := extractInstanceInfo(instance)

	// Assert result
	assert.Equal(t, "i-12345", result.ID)
	assert.Equal(t, "test-instance", result.Name)
	assert.Equal(t, "t2.micro", result.Type)
	assert.Equal(t, "running", result.State)
	assert.Equal(t, "1.2.3.4", result.PublicIP)
	assert.Equal(t, "10.0.0.1", result.PrivateIP)
	assert.Equal(t, "us-east-1a", result.AZ)
	assert.Equal(t, "vpc-12345", result.VpcID)
	assert.Equal(t, "subnet-12345", result.SubnetID)
	assert.Equal(t, "test", result.Tags["Environment"])
}
