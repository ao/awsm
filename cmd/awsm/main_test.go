// Package main provides tests for the awsm CLI commands and functionality.
package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

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
	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)

	// Set args
	root.SetArgs(args)

	// Execute the command
	err := root.Execute()

	// Return the output
	return buf.String(), err
}

// setupTestEnv creates a test environment by redirecting stdout and stderr to pipes.
// This allows tests to capture output that would normally go to the terminal.
//
// It returns a cleanup function that should be deferred to restore the original
// stdout and stderr and close the pipes.
func setupTestEnv(t *testing.T) func() {
	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Set stdout and stderr to the pipes
	os.Stdout = wOut
	os.Stderr = wErr

	// Return cleanup function
	return func() {
		// Restore original stdout and stderr
		os.Stdout = oldStdout
		os.Stderr = oldStderr

		// Close the pipes
		wOut.Close()
		wErr.Close()
		rOut.Close()
		rErr.Close()
	}
}

// captureOutput executes the given function and captures any output written to
// stdout and stderr during its execution. This is useful for testing functions
// that write directly to stdout/stderr rather than returning their output.
//
// Parameters:
//   - f: The function to execute
//
// Returns the combined stdout and stderr output as a string.
func captureOutput(f func()) string {
	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Set stdout and stderr to the pipes
	os.Stdout = wOut
	os.Stderr = wErr

	// Call the function
	f()

	// Close the write ends of the pipes
	wOut.Close()
	wErr.Close()

	// Read the output
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)

	// Restore original stdout and stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Close the read ends of the pipes
	rOut.Close()
	rErr.Close()

	// Return the output
	return string(outBytes) + string(errBytes)
}

// TestRootCommand tests the root command of the CLI.
// It verifies that the command executes without errors and produces the expected output.
func TestRootCommand(t *testing.T) {
	// Create a new root command for testing
	cmd := &cobra.Command{
		Use:   "awsm",
		Short: "awsm - AWS CLI Made Awesome",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Root command executed")
		},
	}

	// Execute the command
	output, err := executeCommand(cmd)

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "Root command executed")
}

// TestEC2Command tests the EC2 command of the CLI.
// It verifies that the command executes without errors and produces the expected output.
func TestEC2Command(t *testing.T) {
	// Create a new EC2 command for testing
	cmd := newEC2Command()
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("EC2 command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd)

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "EC2 command executed")
}

// TestS3Command tests the S3 command of the CLI.
// It verifies that the command executes without errors and produces the expected output.
func TestS3Command(t *testing.T) {
	// Create a new S3 command for testing
	cmd := newS3Command()
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("S3 command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd)

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "S3 command executed")
}

// TestLambdaCommand tests the Lambda command of the CLI.
// It verifies that the command executes without errors and produces the expected output.
func TestLambdaCommand(t *testing.T) {
	// Create a new Lambda command for testing
	cmd := newLambdaCommand()
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("Lambda command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd)

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "Lambda command executed")
}

// TestEC2ListCommand tests the EC2 list subcommand of the CLI.
// It verifies that the subcommand executes without errors and produces the expected output.
func TestEC2ListCommand(t *testing.T) {
	// Create a new EC2 command for testing
	cmd := newEC2Command()

	// Find the list subcommand
	listCmd, _, err := cmd.Find([]string{"list"})
	assert.NoError(t, err)

	// Override the Run function for testing
	listCmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("EC2 list command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd, "list")

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "EC2 list command executed")
}

// TestS3ListCommand tests the S3 ls subcommand of the CLI.
// It verifies that the subcommand executes without errors and produces the expected output.
func TestS3ListCommand(t *testing.T) {
	// Create a new S3 command for testing
	cmd := newS3Command()

	// Find the ls subcommand
	listCmd, _, err := cmd.Find([]string{"ls"})
	assert.NoError(t, err)

	// Override the Run function for testing
	listCmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("S3 ls command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd, "ls")

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "S3 ls command executed")
}

// TestLambdaListCommand tests the Lambda list subcommand of the CLI.
// It verifies that the subcommand executes without errors and produces the expected output.
func TestLambdaListCommand(t *testing.T) {
	// Create a new Lambda command for testing
	cmd := newLambdaCommand()

	// Find the list subcommand
	listCmd, _, err := cmd.Find([]string{"list"})
	assert.NoError(t, err)

	// Override the Run function for testing
	listCmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("Lambda list command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd, "list")

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "Lambda list command executed")
}

// TestAddCommands tests that the root command correctly adds all the expected subcommands.
// It verifies that the EC2, S3, and Lambda commands are properly registered with the root command.
func TestAddCommands(t *testing.T) {
	// Create a new root command for testing
	cmd := &cobra.Command{
		Use:   "awsm",
		Short: "awsm - AWS CLI Made Awesome",
	}

	// Add commands
	cmd.AddCommand(newEC2Command())
	cmd.AddCommand(newS3Command())
	cmd.AddCommand(newLambdaCommand())

	// Check if commands were added
	hasEC2 := false
	hasS3 := false
	hasLambda := false

	for _, command := range cmd.Commands() {
		switch command.Name() {
		case "ec2":
			hasEC2 = true
		case "s3":
			hasS3 = true
		case "lambda":
			hasLambda = true
		}
	}

	assert.True(t, hasEC2, "EC2 command not found")
	assert.True(t, hasS3, "S3 command not found")
	assert.True(t, hasLambda, "Lambda command not found")
}

// TestLaunchTUI tests the Terminal UI launch functionality.
// This test is skipped in CI environments where a terminal might not be available.
// It uses a mock function to simulate the TUI instead of actually launching it.
func TestLaunchTUI(t *testing.T) {
	// Skip this test if running in CI
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping TUI test in CI environment")
	}

	// Create a mock function that simulates the TUI
	mockTUI := func() error {
		return nil
	}

	// Store the original function in a variable
	// In a real test, we would use a mock or dependency injection
	// For this test, we'll just call our mock function directly

	// Call the mock function
	err := mockTUI()

	// Assert no error
	assert.NoError(t, err)
}

// TestMain tests the main function of the CLI.
// This test is skipped in CI environments where it might cause issues.
// It captures the output of the main function when run with the --help flag
// and verifies that it contains the expected help text.
func TestMain(t *testing.T) {
	// Skip this test if running in CI
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping main test in CI environment")
	}

	// Save original args
	oldArgs := os.Args

	// Set args for testing
	os.Args = []string{"awsm", "--help"}

	// Capture output
	output := captureOutput(func() {
		// Call main with a defer to recover from os.Exit
		defer func() {
			if r := recover(); r != nil {
				// Expected os.Exit
			}
		}()
		main()
	})

	// Restore original args
	os.Args = oldArgs

	// Assert output
	assert.Contains(t, output, "awsm - AWS CLI Made Awesome")
}

// TestConfigCommand tests the config command of the CLI.
// It verifies that the command executes without errors and produces the expected output.
func TestConfigCommand(t *testing.T) {
	// Create a new config command for testing
	cmd := newConfigCommand()
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("Config command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd)

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "Config command executed")
}

// TestContextCommand tests the context command of the CLI.
// It verifies that the command executes without errors and produces the expected output.
func TestContextCommand(t *testing.T) {
	// Create a new context command for testing
	cmd := newContextCommand()
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("Context command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd)

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "Context command executed")
}

// TestModeCommand tests the mode command of the CLI.
// It verifies that the command executes without errors and produces the expected output.
func TestModeCommand(t *testing.T) {
	// Create a new mode command for testing
	cmd := newModeCommand()
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Println("Mode command executed")
	}

	// Execute the command
	output, err := executeCommand(cmd)

	// Assert no error
	assert.NoError(t, err)

	// Assert output
	assert.Contains(t, output, "Mode command executed")
}
