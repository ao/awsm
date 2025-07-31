package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ao/awsm/internal/aws/ec2"
	"github.com/ao/awsm/internal/aws/lambda"
	"github.com/ao/awsm/internal/aws/s3"
	"github.com/ao/awsm/internal/config"
	"github.com/ao/awsm/internal/logger"
	"github.com/ao/awsm/internal/tui"
	"github.com/ao/awsm/internal/utils"
	"github.com/spf13/cobra"
)

// Import version information
var (
	// Version is imported from the root package
	Version = "0.1.0"
	// BuildTime is imported from the root package
	BuildTime = "unknown"
	// CommitHash is imported from the root package
	CommitHash = "unknown"
)

var (
	// Global flags
	awsProfile   string
	awsRegion    string
	outputFormat string
	tuiMode      bool

	// Root command
	rootCmd = &cobra.Command{
		Use:   "awsm",
		Short: "awsm - AWS CLI Made Awesome",
		Long: `awsm is a tool designed to make AWS CLI commands easier and more pleasant to use.
		
It provides a more intuitive interface to AWS services with enhanced features:
- Simplified command structure
- Rich output formatting
- Interactive TUI mode
- Profile and region management
- Improved error messages`,
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip for help and version commands
			if cmd.Name() == "help" || cmd.Name() == "version" {
				return nil
			}

			// Initialize configuration
			if err := config.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize configuration: %w", err)
			}

			// Check if context flag is provided
			if contextName, _ := cmd.Flags().GetString("context"); contextName != "" {
				if err := config.SetCurrentContext(contextName); err != nil {
					return fmt.Errorf("failed to set context: %w", err)
				}
			} else {
				// Update configuration with flag values if provided
				if awsProfile != "" {
					if err := config.SetAWSProfile(awsProfile); err != nil {
						return fmt.Errorf("failed to set AWS profile: %w", err)
					}
				}

				if awsRegion != "" {
					if err := config.SetAWSRegion(awsRegion); err != nil {
						return fmt.Errorf("failed to set AWS region: %w", err)
					}
				}
			}

			if outputFormat != "" {
				if !utils.IsValidOutputFormat(outputFormat) {
					return fmt.Errorf("invalid output format: %s", outputFormat)
				}
				if err := config.SetOutputFormat(outputFormat); err != nil {
					return fmt.Errorf("failed to set output format: %w", err)
				}
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to awsm - AWS CLI Made Awesome!")
			fmt.Printf("Version: %s\n", Version)
			fmt.Println("\nCurrent Settings:")
			fmt.Printf("  Current Context: %s\n", config.GetCurrentContext())
			fmt.Printf("  AWS Profile: %s\n", config.GetAWSProfile())
			fmt.Printf("  AWS Region: %s\n", config.GetAWSRegion())
			fmt.Printf("  Output Format: %s\n", config.GetOutputFormat())
			fmt.Println("\nUse --help for more information about available commands.")
		},
	}
)

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&awsProfile, "profile", "", "AWS profile to use")
	rootCmd.PersistentFlags().StringVar(&awsRegion, "region", "", "AWS region to use")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "", "Output format (json, yaml, table, text)")
	rootCmd.PersistentFlags().BoolVar(&tuiMode, "tui", false, "Start in TUI mode")
	rootCmd.PersistentFlags().String("context", "", "AWS context to use")

	// Add commands
	addCommands()
}

// addCommands adds all child commands to the root command
func addCommands() {
	// Add service commands
	rootCmd.AddCommand(newEC2Command())
	rootCmd.AddCommand(newS3Command())
	rootCmd.AddCommand(newLambdaCommand())

	// Add mode command
	rootCmd.AddCommand(newModeCommand())

	// Add config command
	rootCmd.AddCommand(newConfigCommand())

	// Add context command
	rootCmd.AddCommand(newContextCommand())

	// Add direct TUI command
	rootCmd.AddCommand(newTUICommand())
}

func main() {
	// If --tui flag is provided, start in TUI mode
	if tuiMode {
		if err := launchTUI(); err != nil {
			utils.PrintError(err)
			os.Exit(1)
		}
		return
	}

	// Otherwise, execute the CLI command
	if err := rootCmd.Execute(); err != nil {
		utils.PrintError(err)
		os.Exit(1)
	}
}

// launchTUI launches the TUI application
func launchTUI() error {
	// Initialize the logger
	if err := logger.Initialize(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		// Continue even if logger fails, as it's not critical
	}
	defer logger.Close()

	// Log version information
	logger.Info("Launching TUI with Version=%s, BuildTime=%s, CommitHash=%s",
		Version, BuildTime, CommitHash)

	// Create a debug file to verify this function is being called
	debugFile, err := os.Create("launch_debug.log")
	if err == nil {
		debugFile.WriteString(fmt.Sprintf("Version: %s\nBuildTime: %s\nCommitHash: %s\n",
			Version, BuildTime, CommitHash))
		debugFile.Close()
	}

	// Pass version information to the TUI package
	tui.SetVersionInfo(Version, BuildTime, CommitHash)

	// Run the TUI application
	return tui.Run()
}

// newEC2Command creates the ec2 command
func newEC2Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ec2",
		Short: "EC2 instance management",
		Long:  `Manage EC2 instances, security groups, and related resources.`,
	}

	// Add subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List EC2 instances",
			Long:  `List EC2 instances with optional filtering.`,
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()

				// Create EC2 adapter
				adapter, err := ec2.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create EC2 adapter: %w", err))
					return
				}

				// List EC2 instances
				instances, err := adapter.ListInstances(ctx, nil, 0)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to list EC2 instances: %w", err))
					return
				}

				// Format and print the output
				utils.PrintOutput(instances, config.GetOutputFormat())
			},
		},
		&cobra.Command{
			Use:   "describe [instance-id]",
			Short: "Describe an EC2 instance",
			Long:  `Show detailed information about an EC2 instance.`,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()
				instanceID := args[0]

				// Create EC2 adapter
				adapter, err := ec2.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create EC2 adapter: %w", err))
					return
				}

				// Describe EC2 instance
				instance, err := adapter.DescribeInstance(ctx, instanceID)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to describe EC2 instance %s: %w", instanceID, err))
					return
				}

				// Format and print the output
				utils.PrintOutput(instance, config.GetOutputFormat())
			},
		},
		&cobra.Command{
			Use:   "start [instance-id]",
			Short: "Start an EC2 instance",
			Long:  `Start a stopped EC2 instance.`,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()
				instanceID := args[0]

				// Create EC2 adapter
				adapter, err := ec2.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create EC2 adapter: %w", err))
					return
				}

				// Start EC2 instance
				if err := adapter.StartInstance(ctx, instanceID); err != nil {
					utils.PrintError(fmt.Errorf("failed to start EC2 instance %s: %w", instanceID, err))
					return
				}

				fmt.Printf("Successfully started EC2 instance %s\n", instanceID)
			},
		},
		&cobra.Command{
			Use:   "stop [instance-id]",
			Short: "Stop an EC2 instance",
			Long:  `Stop a running EC2 instance.`,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()
				instanceID := args[0]

				// Create EC2 adapter
				adapter, err := ec2.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create EC2 adapter: %w", err))
					return
				}

				// Stop EC2 instance
				if err := adapter.StopInstance(ctx, instanceID); err != nil {
					utils.PrintError(fmt.Errorf("failed to stop EC2 instance %s: %w", instanceID, err))
					return
				}

				fmt.Printf("Successfully stopped EC2 instance %s\n", instanceID)
			},
		},
	)

	return cmd
}

// newS3Command creates the s3 command
func newS3Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "S3 bucket and object management",
		Long:  `Manage S3 buckets, objects, and related resources.`,
	}

	// Add subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "ls [bucket-name]",
			Short: "List S3 buckets or objects",
			Long:  `List S3 buckets or objects in a bucket.`,
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()

				// Create S3 adapter
				adapter, err := s3.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create S3 adapter: %w", err))
					return
				}

				if len(args) == 0 {
					// List S3 buckets
					buckets, err := adapter.ListBuckets(ctx)
					if err != nil {
						utils.PrintError(fmt.Errorf("failed to list S3 buckets: %w", err))
						return
					}

					// Format and print the output
					utils.PrintOutput(buckets, config.GetOutputFormat())
				} else {
					// List objects in bucket
					bucketName := args[0]
					objects, err := adapter.ListObjects(ctx, bucketName, "", 0)
					if err != nil {
						utils.PrintError(fmt.Errorf("failed to list objects in bucket %s: %w", bucketName, err))
						return
					}

					// Format and print the output
					utils.PrintOutput(objects, config.GetOutputFormat())
				}
			},
		},
		&cobra.Command{
			Use:   "cp [source] [destination]",
			Short: "Copy objects to/from S3",
			Long:  `Copy objects to or from S3 buckets.`,
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()
				source := args[0]
				destination := args[1]

				// Create S3 adapter
				adapter, err := s3.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create S3 adapter: %w", err))
					return
				}

				// Check if source is an S3 URL (s3://bucket/key)
				if strings.HasPrefix(source, "s3://") {
					// Download from S3
					parts := strings.SplitN(strings.TrimPrefix(source, "s3://"), "/", 2)
					if len(parts) != 2 {
						utils.PrintError(fmt.Errorf("invalid S3 URL: %s", source))
						return
					}

					bucketName := parts[0]
					key := parts[1]

					if err := adapter.DownloadObject(ctx, bucketName, key, destination); err != nil {
						utils.PrintError(fmt.Errorf("failed to download object: %w", err))
						return
					}

					fmt.Printf("Downloaded s3://%s/%s to %s\n", bucketName, key, destination)
				} else {
					// Upload to S3
					parts := strings.SplitN(strings.TrimPrefix(destination, "s3://"), "/", 2)
					if len(parts) != 2 {
						utils.PrintError(fmt.Errorf("invalid S3 URL: %s", destination))
						return
					}

					bucketName := parts[0]
					key := parts[1]

					if err := adapter.UploadObject(ctx, bucketName, key, source); err != nil {
						utils.PrintError(fmt.Errorf("failed to upload object: %w", err))
						return
					}

					fmt.Printf("Uploaded %s to s3://%s/%s\n", source, bucketName, key)
				}
			},
		},
		&cobra.Command{
			Use:   "rm [bucket-name/object-key]",
			Short: "Remove an S3 object",
			Long:  `Remove an object from an S3 bucket.`,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()
				s3Path := args[0]

				// Create S3 adapter
				adapter, err := s3.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create S3 adapter: %w", err))
					return
				}

				// Parse the S3 path
				parts := strings.SplitN(strings.TrimPrefix(s3Path, "s3://"), "/", 2)
				if len(parts) != 2 {
					utils.PrintError(fmt.Errorf("invalid S3 path: %s", s3Path))
					return
				}

				bucketName := parts[0]
				key := parts[1]

				// Delete the object
				if err := adapter.DeleteObject(ctx, bucketName, key); err != nil {
					utils.PrintError(fmt.Errorf("failed to delete object: %w", err))
					return
				}

				fmt.Printf("Removed s3://%s/%s\n", bucketName, key)
			},
		},
	)

	return cmd
}

// newLambdaCommand creates the lambda command
func newLambdaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lambda",
		Short: "Lambda function management",
		Long:  `Manage Lambda functions, layers, and related resources.`,
	}

	// Add subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List Lambda functions",
			Long:  `List Lambda functions with optional filtering.`,
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()

				// Create Lambda adapter
				adapter, err := lambda.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create Lambda adapter: %w", err))
					return
				}

				// List Lambda functions
				functions, err := adapter.ListFunctions(ctx, 0)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to list Lambda functions: %w", err))
					return
				}

				// Format and print the output
				utils.PrintOutput(functions, config.GetOutputFormat())
			},
		},
		&cobra.Command{
			Use:   "invoke [function-name]",
			Short: "Invoke a Lambda function",
			Long:  `Invoke a Lambda function and display the result.`,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()
				functionName := args[0]

				// Create Lambda adapter
				adapter, err := lambda.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create Lambda adapter: %w", err))
					return
				}

				// Create empty payload
				payload, err := lambda.FormatPayload(map[string]interface{}{})
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to format payload: %w", err))
					return
				}

				// Invoke Lambda function
				result, err := adapter.InvokeFunction(ctx, functionName, payload)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to invoke Lambda function %s: %w", functionName, err))
					return
				}

				// Check for function error
				if result.FunctionError != "" {
					utils.PrintError(fmt.Errorf("function execution error: %s", result.FunctionError))
					return
				}

				// Format and print the output
				var responseData interface{}
				if err := lambda.ParsePayload(result.Payload, &responseData); err != nil {
					utils.PrintError(fmt.Errorf("failed to parse response: %w", err))
					return
				}

				utils.PrintOutput(responseData, config.GetOutputFormat())
			},
		},
		&cobra.Command{
			Use:   "logs [function-name]",
			Short: "Show logs for a Lambda function",
			Long:  `Display CloudWatch logs for a Lambda function.`,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				ctx := context.Background()
				functionName := args[0]

				// Create Lambda adapter
				adapter, err := lambda.NewAdapter(ctx)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to create Lambda adapter: %w", err))
					return
				}

				// Get logs for Lambda function (last 100 events)
				logs, err := adapter.GetFunctionLogs(ctx, functionName, time.Time{}, 100)
				if err != nil {
					utils.PrintError(fmt.Errorf("failed to get logs for Lambda function %s: %w", functionName, err))
					return
				}

				// Format and print the output
				utils.PrintOutput(logs, config.GetOutputFormat())
			},
		},
	)

	return cmd
}

// newModeCommand creates the mode command for switching between CLI and TUI modes
func newModeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mode [cli|tui]",
		Short: "Switch between CLI and TUI modes",
		Long:  `Switch between command-line interface (CLI) and terminal user interface (TUI) modes.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := args[0]
			if mode != "cli" && mode != "tui" {
				return fmt.Errorf("invalid mode: %s (must be 'cli' or 'tui')", mode)
			}

			if err := config.SetAppMode(mode); err != nil {
				return fmt.Errorf("failed to set mode: %w", err)
			}

			fmt.Printf("Switched to %s mode\n", mode)

			if mode == "tui" {
				// Launch the TUI
				return launchTUI()
			}

			return nil
		},
	}

	return cmd
}

// newConfigCommand creates the config command for managing configuration
func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `View and modify configuration settings.`,
	}

	// Add subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "get [key]",
			Short: "Get a configuration value",
			Long:  `Get the value of a configuration setting.`,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				key := args[0]
				switch key {
				case "profile":
					fmt.Println(config.GetAWSProfile())
				case "region":
					fmt.Println(config.GetAWSRegion())
				case "output":
					fmt.Println(config.GetOutputFormat())
				case "mode":
					fmt.Println(config.GetAppMode())
				default:
					fmt.Printf("Unknown configuration key: %s\n", key)
				}
			},
		},
		&cobra.Command{
			Use:   "set [key] [value]",
			Short: "Set a configuration value",
			Long:  `Set the value of a configuration setting.`,
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				key := args[0]
				value := args[1]

				var err error
				switch key {
				case "profile":
					err = config.SetAWSProfile(value)
				case "region":
					err = config.SetAWSRegion(value)
				case "output":
					if !utils.IsValidOutputFormat(value) {
						return fmt.Errorf("invalid output format: %s", value)
					}
					err = config.SetOutputFormat(value)
				case "mode":
					if value != "cli" && value != "tui" {
						return fmt.Errorf("invalid mode: %s (must be 'cli' or 'tui')", value)
					}
					err = config.SetAppMode(value)
				default:
					return fmt.Errorf("unknown configuration key: %s", key)
				}

				if err != nil {
					return fmt.Errorf("failed to set %s: %w", key, err)
				}

				fmt.Printf("Set %s to %s\n", key, value)
				return nil
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List all configuration values",
			Long:  `List all configuration settings and their values.`,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Configuration:")
				fmt.Printf("  profile: %s\n", config.GetAWSProfile())
				fmt.Printf("  region: %s\n", config.GetAWSRegion())
				fmt.Printf("  output: %s\n", config.GetOutputFormat())
				fmt.Printf("  mode: %s\n", config.GetAppMode())
			},
		},
	)

	return cmd
}

// newContextCommand creates the context command for managing AWS contexts
func newContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage AWS contexts",
		Long:  `Create, switch, and manage AWS contexts (combinations of profile, region, and role).`,
	}

	// Add subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available contexts",
			Long:  `List all available AWS contexts.`,
			Run: func(cmd *cobra.Command, args []string) {
				// Get contexts
				contexts := config.ListContexts()

				// Format output based on format
				switch config.GetOutputFormat() {
				case "json", "yaml":
					utils.PrintOutput(contexts, config.GetOutputFormat())
				default:
					// Print table format
					fmt.Println("Available Contexts:")
					fmt.Println("-------------------")
					fmt.Printf("%-20s %-15s %-15s %-30s\n", "NAME", "PROFILE", "REGION", "ROLE")
					for _, ctx := range contexts {
						current := " "
						if ctx.Current {
							current = "*"
						}
						fmt.Printf("%s %-19s %-15s %-15s %-30s\n",
							current, ctx.Name, ctx.Profile, ctx.Region, ctx.Role)
					}
				}
			},
		},
		&cobra.Command{
			Use:   "current",
			Short: "Show current context",
			Long:  `Display information about the current AWS context.`,
			Run: func(cmd *cobra.Command, args []string) {
				// Get current context
				ctx, err := config.GetCurrentContextInfo()
				if err != nil {
					utils.PrintError(err)
					return
				}

				// Format output based on format
				switch config.GetOutputFormat() {
				case "json", "yaml":
					utils.PrintOutput(ctx, config.GetOutputFormat())
				default:
					fmt.Println("Current Context:")
					fmt.Printf("  Name:    %s\n", ctx.Name)
					fmt.Printf("  Profile: %s\n", ctx.Profile)
					fmt.Printf("  Region:  %s\n", ctx.Region)
					if ctx.Role != "" {
						fmt.Printf("  Role:    %s\n", ctx.Role)
					}
				}
			},
		},
		&cobra.Command{
			Use:   "use [context-name]",
			Short: "Switch to a different context",
			Long:  `Switch to a different AWS context.`,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				contextName := args[0]

				// Switch context
				if err := config.SwitchContext(contextName); err != nil {
					return fmt.Errorf("failed to switch context: %w", err)
				}

				fmt.Printf("Switched to context '%s'\n", contextName)
				return nil
			},
		},
		&cobra.Command{
			Use:   "create [context-name]",
			Short: "Create a new context",
			Long:  `Create a new AWS context with specified profile, region, and optional role.`,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				contextName := args[0]

				// Get flags
				profile, _ := cmd.Flags().GetString("profile")
				region, _ := cmd.Flags().GetString("region")
				role, _ := cmd.Flags().GetString("role")

				// Validate required flags
				if profile == "" {
					return fmt.Errorf("profile is required")
				}
				if region == "" {
					return fmt.Errorf("region is required")
				}

				// Create context
				if err := config.NewContext(contextName, profile, region, role); err != nil {
					return fmt.Errorf("failed to create context: %w", err)
				}

				fmt.Printf("Created context '%s'\n", contextName)
				return nil
			},
		},
		&cobra.Command{
			Use:   "delete [context-name]",
			Short: "Delete a context",
			Long:  `Delete an AWS context.`,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				contextName := args[0]

				// Delete context
				if err := config.RemoveContext(contextName); err != nil {
					return fmt.Errorf("failed to delete context: %w", err)
				}

				fmt.Printf("Deleted context '%s'\n", contextName)
				return nil
			},
		},
		&cobra.Command{
			Use:   "import",
			Short: "Import contexts from AWS config",
			Long:  `Import contexts from the AWS config file.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				// Import contexts
				count, err := config.ImportContextsFromAWS()
				if err != nil {
					return fmt.Errorf("failed to import contexts: %w", err)
				}

				fmt.Printf("Imported %d contexts from AWS config\n", count)
				return nil
			},
		},
		&cobra.Command{
			Use:   "export",
			Short: "Export contexts to AWS config",
			Long:  `Export contexts to the AWS config file.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				// Get flags
				overwrite, _ := cmd.Flags().GetBool("overwrite")

				// Export contexts
				if err := config.ExportContextsToAWS(overwrite); err != nil {
					return fmt.Errorf("failed to export contexts: %w", err)
				}

				fmt.Println("Exported contexts to AWS config")
				return nil
			},
		},
	)

	// Add flags to create command
	createCmd := cmd.Commands()[3] // The create command
	createCmd.Flags().String("profile", "", "AWS profile to use")
	createCmd.Flags().String("region", "", "AWS region to use")
	createCmd.Flags().String("role", "", "AWS role to assume (optional)")

	// Add flags to export command
	exportCmd := cmd.Commands()[6] // The export command
	exportCmd.Flags().Bool("overwrite", false, "Overwrite existing AWS config file")

	return cmd
}

// newTUICommand creates a direct command to launch the TUI
func newTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the Terminal User Interface",
		Long:  `Launch the Terminal User Interface (TUI) for interactive AWS management.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := launchTUI(); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
