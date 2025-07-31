// Package lambda provides functionality for interacting with AWS Lambda functions.
// It includes operations for listing functions, getting function details,
// invoking functions, and retrieving function logs from CloudWatch.
package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ao/awsm/internal/aws/client"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaClient defines the interface for Lambda client operations.
// This interface allows for easy mocking in tests.
type LambdaClient interface {
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
	GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

// CloudWatchLogsClient defines the interface for CloudWatch Logs client operations.
// This interface allows for easy mocking in tests and is used to retrieve
// logs for Lambda function executions.
type CloudWatchLogsClient interface {
	FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
}

// Adapter represents a Lambda service adapter that provides
// higher-level operations for interacting with Lambda functions.
type Adapter struct {
	client     LambdaClient         // AWS Lambda client implementation
	logsClient CloudWatchLogsClient // AWS CloudWatch Logs client for retrieving function logs
}

// Function represents a Lambda function with relevant information.
// This is a simplified representation of the AWS Lambda function configuration
// that includes only the most commonly used fields.
type Function struct {
	Name         string            // Name of the Lambda function
	Description  string            // Description of the function
	Runtime      string            // Runtime environment (e.g., nodejs14.x, python3.9)
	Handler      string            // Function handler (e.g., index.handler)
	Role         string            // IAM role ARN used by the function
	Size         int64             // Size of the function code in bytes
	Timeout      int32             // Function timeout in seconds
	Memory       int32             // Memory allocation in MB
	LastModified string            // When the function was last modified
	Version      string            // Function version
	Environment  map[string]string // Environment variables
	Tags         map[string]string // Function tags
}

// LogEvent represents a CloudWatch log event from a Lambda function execution.
type LogEvent struct {
	Timestamp int64  // Unix timestamp in milliseconds
	Message   string // Log message content
}

// InvokeResult represents the result of a Lambda function invocation.
// It contains the response from the function execution and any associated metadata.
type InvokeResult struct {
	StatusCode    int32  // HTTP status code of the invocation
	LogResult     string // Base64-encoded last 4KB of execution log
	Payload       []byte // Function execution result payload
	FunctionError string // Error type if the function execution failed
}

// NewAdapter creates a new Lambda adapter using the AWS credentials
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

	// Create Lambda client
	lambdaClient := lambda.NewFromConfig(awsClient.Config)

	// Create CloudWatch Logs client for retrieving Lambda logs
	logsClient := cloudwatchlogs.NewFromConfig(awsClient.Config)

	return &Adapter{
		client:     lambdaClient,
		logsClient: logsClient,
	}, nil
}

// NewAdapterWithClients creates a new Lambda adapter with provided clients.
// This is particularly useful for testing with mock clients.
func NewAdapterWithClients(lambdaClient LambdaClient, logsClient CloudWatchLogsClient) *Adapter {
	return &Adapter{
		client:     lambdaClient,
		logsClient: logsClient,
	}
}

// ListFunctions lists Lambda functions with optional maximum item limit.
//
// Parameters:
//   - ctx: Context for the API call
//   - maxItems: Maximum number of functions to return (0 for no limit)
//
// Returns a slice of Function structs and an error if the operation fails.
func (a *Adapter) ListFunctions(ctx context.Context, maxItems int32) ([]Function, error) {
	// Create the input for the ListFunctions API
	input := &lambda.ListFunctionsInput{}

	// Create paginator
	paginator := lambda.NewListFunctionsPaginator(a.client, input)

	var functions []Function
	var count int32 = 0

	// Iterate through pages
	for paginator.HasMorePages() && (maxItems == 0 || count < maxItems) {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Lambda functions: %w", err)
		}

		// Process each function
		for _, function := range output.Functions {
			// Skip if we've reached the maximum number of items
			if maxItems > 0 && count >= maxItems {
				break
			}

			// Extract function information
			fn := extractFunctionInfo(function)
			functions = append(functions, fn)
			count++
		}
	}

	return functions, nil
}

// GetFunction gets detailed information about a specific Lambda function.
//
// Parameters:
//   - ctx: Context for the API call
//   - functionName: The name or ARN of the Lambda function
//
// Returns a pointer to a Function struct with the function details
// or an error if the function cannot be found or retrieved.
func (a *Adapter) GetFunction(ctx context.Context, functionName string) (*Function, error) {
	// Create the input for the GetFunction API
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}

	// Call the GetFunction API
	output, err := a.client.GetFunction(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda function %s: %w", functionName, err)
	}

	// Extract function information
	function := extractFunctionInfo(*output.Configuration)

	// Add tags if available
	if output.Tags != nil {
		function.Tags = output.Tags
	}

	return &function, nil
}

// InvokeFunction invokes a Lambda function with the provided payload.
//
// Parameters:
//   - ctx: Context for the API call
//   - functionName: The name or ARN of the Lambda function to invoke
//   - payload: The JSON payload to pass to the function (can be nil)
//
// Returns an InvokeResult containing the function's response and execution details,
// or an error if the invocation fails.
func (a *Adapter) InvokeFunction(ctx context.Context, functionName string, payload []byte) (*InvokeResult, error) {
	// Create the input for the Invoke API
	input := &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
		LogType:      types.LogTypeTail, // Include the execution log in the response
	}

	// Call the Invoke API
	output, err := a.client.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda function %s: %w", functionName, err)
	}

	// Create the result
	result := &InvokeResult{
		StatusCode:    output.StatusCode,
		LogResult:     aws.ToString(output.LogResult),
		Payload:       output.Payload,
		FunctionError: aws.ToString(output.FunctionError),
	}

	return result, nil
}

// GetFunctionLogs gets the CloudWatch logs for a Lambda function.
//
// Parameters:
//   - ctx: Context for the API call
//   - functionName: The name of the Lambda function
//   - startTime: The start time for log retrieval (zero value for no start time)
//   - limit: Maximum number of log events to return (0 for no limit)
//
// Returns a slice of LogEvent structs and an error if the operation fails.
func (a *Adapter) GetFunctionLogs(ctx context.Context, functionName string, startTime time.Time, limit int32) ([]LogEvent, error) {
	// Get the log group name for the Lambda function
	logGroupName := fmt.Sprintf("/aws/lambda/%s", functionName)

	// Create the input for the FilterLogEvents API
	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(logGroupName),
		Limit:        aws.Int32(limit),
	}

	// Add start time if provided
	if !startTime.IsZero() {
		startTimeMillis := startTime.UnixNano() / int64(time.Millisecond)
		input.StartTime = aws.Int64(startTimeMillis)
	}

	// Call the FilterLogEvents API
	paginator := cloudwatchlogs.NewFilterLogEventsPaginator(a.logsClient, input)

	var logEvents []LogEvent
	var count int32 = 0

	// Iterate through pages
	for paginator.HasMorePages() && (limit == 0 || count < limit) {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs for Lambda function %s: %w", functionName, err)
		}

		// Process each log event
		for _, event := range output.Events {
			// Skip if we've reached the limit
			if limit > 0 && count >= limit {
				break
			}

			// Extract log event information
			logEvent := LogEvent{
				Timestamp: aws.ToInt64(event.Timestamp),
				Message:   aws.ToString(event.Message),
			}

			logEvents = append(logEvents, logEvent)
			count++
		}
	}

	return logEvents, nil
}

// extractFunctionInfo extracts relevant information from a Lambda function configuration
// and converts it to our simplified Function struct.
//
// This is an internal helper function used by ListFunctions and GetFunction.
func extractFunctionInfo(function types.FunctionConfiguration) Function {
	// Initialize the function
	fn := Function{
		Name:         aws.ToString(function.FunctionName),
		Runtime:      string(function.Runtime),
		Handler:      aws.ToString(function.Handler),
		Role:         aws.ToString(function.Role),
		Size:         function.CodeSize,
		Timeout:      aws.ToInt32(function.Timeout),
		Memory:       aws.ToInt32(function.MemorySize),
		LastModified: aws.ToString(function.LastModified),
		Version:      aws.ToString(function.Version),
		Environment:  make(map[string]string),
		Tags:         make(map[string]string),
	}

	// Extract description if available
	if function.Description != nil {
		fn.Description = *function.Description
	}

	// Extract environment variables if available
	if function.Environment != nil && function.Environment.Variables != nil {
		fn.Environment = function.Environment.Variables
	}

	return fn
}

// FormatPayload formats a Go data structure into a JSON payload
// suitable for Lambda function invocation.
//
// Parameters:
//   - data: The Go data structure to convert to JSON
//
// Returns the JSON byte array and an error if the marshaling fails.
func FormatPayload(data interface{}) ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to format payload: %w", err)
	}
	return payload, nil
}

// ParsePayload parses a Lambda function invocation result payload into
// a provided Go data structure.
//
// Parameters:
//   - payload: The JSON payload from the Lambda function response
//   - v: A pointer to the Go data structure to populate
//
// Returns an error if the unmarshaling fails.
func ParsePayload(payload []byte, v interface{}) error {
	if err := json.Unmarshal(payload, v); err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}
	return nil
}
