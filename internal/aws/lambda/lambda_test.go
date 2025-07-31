// Package lambda provides tests for the Lambda adapter functionality.
package lambda

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogsTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockLambdaClient implements the LambdaClient interface for testing purposes.
// It uses the testify/mock package to mock AWS Lambda API calls.
type mockLambdaClient struct {
	mock.Mock
}

func (m *mockLambdaClient) ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*lambda.ListFunctionsOutput), args.Error(1)
}

func (m *mockLambdaClient) GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*lambda.GetFunctionOutput), args.Error(1)
}

func (m *mockLambdaClient) Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*lambda.InvokeOutput), args.Error(1)
}

// mockCloudWatchLogsClient implements the CloudWatchLogsClient interface for testing purposes.
// It uses the testify/mock package to mock AWS CloudWatch Logs API calls.
type mockCloudWatchLogsClient struct {
	mock.Mock
}

func (m *mockCloudWatchLogsClient) FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*cloudwatchlogs.FilterLogEventsOutput), args.Error(1)
}

// These static assertions verify at compile time that the mock clients implement their respective interfaces.
var _ LambdaClient = (*mockLambdaClient)(nil)
var _ CloudWatchLogsClient = (*mockCloudWatchLogsClient)(nil)

// createMockFunctionConfiguration is a helper function that creates a mock Lambda function configuration
// with the specified parameters. This function simplifies the creation of test data for Lambda function tests.
//
// Parameters:
//   - name: The Lambda function name
//   - description: The function description
//   - runtime: The runtime environment (e.g., nodejs14.x, python3.9)
//   - handler: The function handler (e.g., index.handler)
//   - role: The IAM role ARN
//   - codeSize: The size of the function code in bytes
//   - timeout: The function timeout in seconds
//   - memory: The memory allocation in MB
//   - lastModified: When the function was last modified (formatted string)
//   - version: The function version
//   - env: Environment variables for the function
//
// Returns a types.FunctionConfiguration that can be used in test mocks.
func createMockFunctionConfiguration(name, description, runtime, handler, role string, codeSize int64, timeout, memory int32, lastModified, version string, env map[string]string) types.FunctionConfiguration {
	// Create function configuration
	functionConfig := types.FunctionConfiguration{
		FunctionName: aws.String(name),
		Runtime:      types.Runtime(runtime),
		Handler:      aws.String(handler),
		Role:         aws.String(role),
		CodeSize:     codeSize,
		Timeout:      aws.Int32(timeout),
		MemorySize:   aws.Int32(memory),
		LastModified: aws.String(lastModified),
		Version:      aws.String(version),
	}

	// Add description if provided
	if description != "" {
		functionConfig.Description = aws.String(description)
	}

	// Add environment variables if provided
	if len(env) > 0 {
		functionConfig.Environment = &types.EnvironmentResponse{
			Variables: env,
		}
	}

	return functionConfig
}

// TestListFunctions tests the ListFunctions method of the Lambda Adapter.
// It verifies that the adapter correctly processes the AWS API response
// and returns the expected list of functions with all fields properly populated.
func TestListFunctions(t *testing.T) {
	// Create mock clients
	mockLambdaClient := new(mockLambdaClient)
	mockLogsClient := new(mockCloudWatchLogsClient)

	// Create adapter with mock clients
	adapter := NewAdapterWithClients(mockLambdaClient, mockLogsClient)

	// Create mock functions
	function1 := createMockFunctionConfiguration(
		"test-function-1",
		"Test function 1",
		"nodejs14.x",
		"index.handler",
		"arn:aws:iam::123456789012:role/lambda-role",
		1024,
		30,
		128,
		time.Now().Add(-24*time.Hour).Format(time.RFC3339),
		"1",
		map[string]string{"ENV_VAR_1": "value1"},
	)

	function2 := createMockFunctionConfiguration(
		"test-function-2",
		"Test function 2",
		"python3.9",
		"app.handler",
		"arn:aws:iam::123456789012:role/lambda-role",
		2048,
		60,
		256,
		time.Now().Add(-48*time.Hour).Format(time.RFC3339),
		"2",
		map[string]string{"ENV_VAR_2": "value2"},
	)

	// Create mock response
	mockResponse := &lambda.ListFunctionsOutput{
		Functions: []types.FunctionConfiguration{function1, function2},
	}

	// Set up expectations
	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Call the function
	ctx := context.Background()
	functions, err := adapter.ListFunctions(ctx, 0)

	// Assert no error
	assert.NoError(t, err)

	// Assert functions
	assert.Len(t, functions, 2)

	// Assert function 1
	assert.Equal(t, "test-function-1", functions[0].Name)
	assert.Equal(t, "Test function 1", functions[0].Description)
	assert.Equal(t, "nodejs14.x", functions[0].Runtime)
	assert.Equal(t, "index.handler", functions[0].Handler)
	assert.Equal(t, "arn:aws:iam::123456789012:role/lambda-role", functions[0].Role)
	assert.Equal(t, int64(1024), functions[0].Size)
	assert.Equal(t, int32(30), functions[0].Timeout)
	assert.Equal(t, int32(128), functions[0].Memory)
	assert.Equal(t, "1", functions[0].Version)
	assert.Equal(t, "value1", functions[0].Environment["ENV_VAR_1"])

	// Assert function 2
	assert.Equal(t, "test-function-2", functions[1].Name)
	assert.Equal(t, "Test function 2", functions[1].Description)
	assert.Equal(t, "python3.9", functions[1].Runtime)
	assert.Equal(t, "app.handler", functions[1].Handler)
	assert.Equal(t, "arn:aws:iam::123456789012:role/lambda-role", functions[1].Role)
	assert.Equal(t, int64(2048), functions[1].Size)
	assert.Equal(t, int32(60), functions[1].Timeout)
	assert.Equal(t, int32(256), functions[1].Memory)
	assert.Equal(t, "2", functions[1].Version)
	assert.Equal(t, "value2", functions[1].Environment["ENV_VAR_2"])

	// Verify expectations
	mockLambdaClient.AssertExpectations(t)
}

// TestGetFunction tests the GetFunction method of the Lambda Adapter.
// It verifies that the adapter correctly processes the AWS API response
// and returns the expected function details, including tags.
func TestGetFunction(t *testing.T) {
	// Create mock clients
	mockLambdaClient := new(mockLambdaClient)
	mockLogsClient := new(mockCloudWatchLogsClient)

	// Create adapter with mock clients
	adapter := NewAdapterWithClients(mockLambdaClient, mockLogsClient)

	// Create mock function
	function := createMockFunctionConfiguration(
		"test-function",
		"Test function",
		"nodejs14.x",
		"index.handler",
		"arn:aws:iam::123456789012:role/lambda-role",
		1024,
		30,
		128,
		time.Now().Add(-24*time.Hour).Format(time.RFC3339),
		"1",
		map[string]string{"ENV_VAR": "value"},
	)

	// Create mock tags
	tags := map[string]string{
		"Environment": "test",
		"Project":     "awsm",
	}

	// Create mock response
	mockResponse := &lambda.GetFunctionOutput{
		Configuration: &function,
		Tags:          tags,
	}

	// Set up expectations
	mockLambdaClient.On("GetFunction", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Call the function
	ctx := context.Background()
	result, err := adapter.GetFunction(ctx, "test-function")

	// Assert no error
	assert.NoError(t, err)

	// Assert function
	assert.Equal(t, "test-function", result.Name)
	assert.Equal(t, "Test function", result.Description)
	assert.Equal(t, "nodejs14.x", result.Runtime)
	assert.Equal(t, "index.handler", result.Handler)
	assert.Equal(t, "arn:aws:iam::123456789012:role/lambda-role", result.Role)
	assert.Equal(t, int64(1024), result.Size)
	assert.Equal(t, int32(30), result.Timeout)
	assert.Equal(t, int32(128), result.Memory)
	assert.Equal(t, "1", result.Version)
	assert.Equal(t, "value", result.Environment["ENV_VAR"])
	assert.Equal(t, "test", result.Tags["Environment"])
	assert.Equal(t, "awsm", result.Tags["Project"])

	// Verify expectations
	mockLambdaClient.AssertExpectations(t)
}

// TestInvokeFunction tests the InvokeFunction method of the Lambda Adapter.
// It verifies that the adapter correctly calls the AWS API with the
// expected parameters and processes the response, including status code,
// logs, and payload.
func TestInvokeFunction(t *testing.T) {
	// Create mock clients
	mockLambdaClient := new(mockLambdaClient)
	mockLogsClient := new(mockCloudWatchLogsClient)

	// Create adapter with mock clients
	adapter := NewAdapterWithClients(mockLambdaClient, mockLogsClient)

	// Create mock payload
	payload := []byte(`{"key": "value"}`)
	responsePayload := []byte(`{"result": "success"}`)

	// Create mock response
	mockResponse := &lambda.InvokeOutput{
		StatusCode:    200,
		LogResult:     aws.String("U1RBUlQgUmVxdWVzdElkOiA0NWJhOTQzYi1mZWY0LTExZTgtOTVmOC02ZmExNGMzMmVkMjAgVmVyc2lvbjogJExBVEVTVAo="),
		Payload:       responsePayload,
		FunctionError: aws.String(""),
	}

	// Set up expectations
	mockLambdaClient.On("Invoke", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Call the function
	ctx := context.Background()
	result, err := adapter.InvokeFunction(ctx, "test-function", payload)

	// Assert no error
	assert.NoError(t, err)

	// Assert result
	assert.Equal(t, int32(200), result.StatusCode)
	assert.Equal(t, "U1RBUlQgUmVxdWVzdElkOiA0NWJhOTQzYi1mZWY0LTExZTgtOTVmOC02ZmExNGMzMmVkMjAgVmVyc2lvbjogJExBVEVTVAo=", result.LogResult)
	assert.Equal(t, responsePayload, result.Payload)
	assert.Equal(t, "", result.FunctionError)

	// Verify expectations
	mockLambdaClient.AssertExpectations(t)
}

// TestGetFunctionLogs tests the GetFunctionLogs method of the Lambda Adapter.
// It verifies that the adapter correctly calls the CloudWatch Logs API with
// the expected parameters and processes the log events.
func TestGetFunctionLogs(t *testing.T) {
	// Create mock clients
	mockLambdaClient := new(mockLambdaClient)
	mockLogsClient := new(mockCloudWatchLogsClient)

	// Create adapter with mock clients
	adapter := NewAdapterWithClients(mockLambdaClient, mockLogsClient)

	// Create mock log events
	timestamp1 := time.Now().Add(-1*time.Hour).UnixNano() / int64(time.Millisecond)
	timestamp2 := time.Now().Add(-2*time.Hour).UnixNano() / int64(time.Millisecond)

	// Create mock response
	mockResponse := &cloudwatchlogs.FilterLogEventsOutput{
		Events: []cloudwatchlogsTypes.FilteredLogEvent{
			{
				Timestamp: aws.Int64(timestamp1),
				Message:   aws.String("Log message 1"),
			},
			{
				Timestamp: aws.Int64(timestamp2),
				Message:   aws.String("Log message 2"),
			},
		},
	}

	// Set up expectations
	mockLogsClient.On("FilterLogEvents", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Call the function
	ctx := context.Background()
	startTime := time.Now().Add(-24 * time.Hour)
	logs, err := adapter.GetFunctionLogs(ctx, "test-function", startTime, 10)

	// Assert no error
	assert.NoError(t, err)

	// Assert logs
	assert.Len(t, logs, 2)

	// Assert log 1
	assert.Equal(t, timestamp1, logs[0].Timestamp)
	assert.Equal(t, "Log message 1", logs[0].Message)

	// Assert log 2
	assert.Equal(t, timestamp2, logs[1].Timestamp)
	assert.Equal(t, "Log message 2", logs[1].Message)

	// Verify expectations
	mockLogsClient.AssertExpectations(t)
}

// TestFormatPayload tests the FormatPayload function.
// It verifies that the function correctly converts a Go data structure
// to a JSON payload for Lambda function invocation.
func TestFormatPayload(t *testing.T) {
	// Test data
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	// Call the function
	payload, err := FormatPayload(data)

	// Assert no error
	assert.NoError(t, err)

	// Assert payload
	expectedPayload := []byte(`{"key1":"value1","key2":42,"key3":true}`)
	assert.Equal(t, expectedPayload, payload)
}

// TestParsePayload tests the ParsePayload function.
// It verifies that the function correctly parses a JSON payload
// from a Lambda function response into a Go data structure.
func TestParsePayload(t *testing.T) {
	// Test data
	payload := []byte(`{"key1":"value1","key2":42,"key3":true}`)

	// Call the function
	var result map[string]interface{}
	err := ParsePayload(payload, &result)

	// Assert no error
	assert.NoError(t, err)

	// Assert result
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, float64(42), result["key2"])
	assert.Equal(t, true, result["key3"])
}

// TestExtractFunctionInfo tests the extractFunctionInfo function.
// It verifies that the function correctly extracts information from
// an AWS Lambda function configuration and converts it to our simplified
// Function struct.
func TestExtractFunctionInfo(t *testing.T) {
	// Create mock function
	function := createMockFunctionConfiguration(
		"test-function",
		"Test function",
		"nodejs14.x",
		"index.handler",
		"arn:aws:iam::123456789012:role/lambda-role",
		1024,
		30,
		128,
		time.Now().Add(-24*time.Hour).Format(time.RFC3339),
		"1",
		map[string]string{"ENV_VAR": "value"},
	)

	// Call the function
	result := extractFunctionInfo(function)

	// Assert result
	assert.Equal(t, "test-function", result.Name)
	assert.Equal(t, "Test function", result.Description)
	assert.Equal(t, "nodejs14.x", result.Runtime)
	assert.Equal(t, "index.handler", result.Handler)
	assert.Equal(t, "arn:aws:iam::123456789012:role/lambda-role", result.Role)
	assert.Equal(t, int64(1024), result.Size)
	assert.Equal(t, int32(30), result.Timeout)
	assert.Equal(t, int32(128), result.Memory)
	assert.Equal(t, "1", result.Version)
	assert.Equal(t, "value", result.Environment["ENV_VAR"])
}
