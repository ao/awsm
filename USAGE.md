# AWSM Usage Guide

This guide provides detailed instructions for using AWSM to manage your AWS resources.

## Table of Contents

- [Command Line Interface](#command-line-interface)
  - [Global Flags](#global-flags)
  - [Configuration Commands](#configuration-commands)
  - [Context Commands](#context-commands)
  - [EC2 Commands](#ec2-commands)
  - [S3 Commands](#s3-commands)
  - [Lambda Commands](#lambda-commands)
- [Terminal User Interface (TUI)](#terminal-user-interface-tui)
  - [Navigation](#navigation)
  - [Dashboard](#dashboard)
  - [EC2 View](#ec2-view)
  - [S3 View](#s3-view)
  - [Lambda View](#lambda-view)
  - [Command Palette](#command-palette)
  - [Context Switching](#context-switching)
- [Output Formatting](#output-formatting)
- [Environment Variables](#environment-variables)
- [Configuration File](#configuration-file)
- [Advanced Usage](#advanced-usage)

## Command Line Interface

AWSM provides a command-line interface (CLI) for managing AWS resources. The basic syntax is:

```
awsm [global flags] <command> [subcommand] [flags] [arguments]
```

### Global Flags

- `--profile`, `-p`: AWS profile to use
- `--region`, `-r`: AWS region to use
- `--output`, `-o`: Output format (text, json, yaml)
- `--context`, `-c`: Context to use
- `--verbose`, `-v`: Enable verbose output
- `--help`, `-h`: Show help for a command
- `--version`: Show version information

### Configuration Commands

#### Initialize Configuration

```bash
awsm config init
```

This creates a default configuration file at `~/.awsm.yaml`.

#### Set Configuration Values

```bash
# Set AWS profile
awsm config set aws.profile default

# Set AWS region
awsm config set aws.region us-west-2

# Set output format
awsm config set output.format json
```

#### Get Configuration Values

```bash
# Get AWS profile
awsm config get aws.profile

# Get AWS region
awsm config get aws.region

# Get output format
awsm config get output.format
```

### Context Commands

#### Create a Context

```bash
awsm context create <name> --profile <profile> --region <region> [--role <role>]
```

Example:
```bash
awsm context create dev --profile development --region us-west-2
awsm context create prod --profile production --region us-east-1 --role arn:aws:iam::123456789012:role/admin
```

#### List Contexts

```bash
awsm context list
```

#### Switch Context

```bash
awsm context use <name>
```

Example:
```bash
awsm context use dev
```

#### Show Current Context

```bash
awsm context current
```

#### Update Context

```bash
awsm context update <name> [--profile <profile>] [--region <region>] [--role <role>]
```

Example:
```bash
awsm context update dev --region eu-west-1
```

#### Delete Context

```bash
awsm context delete <name>
```

Example:
```bash
awsm context delete dev
```

### EC2 Commands

#### List EC2 Instances

```bash
awsm ec2 list [--filter <key>=<value>] [--max-items <number>]
```

Example:
```bash
# List all instances
awsm ec2 list

# List instances with a specific tag
awsm ec2 list --filter "tag:Environment=Production"

# List running instances
awsm ec2 list --filter "instance-state-name=running"

# Limit the number of instances returned
awsm ec2 list --max-items 10
```

#### Describe an EC2 Instance

```bash
awsm ec2 describe <instance-id>
```

Example:
```bash
awsm ec2 describe i-1234567890abcdef0
```

#### Start an EC2 Instance

```bash
awsm ec2 start <instance-id>
```

Example:
```bash
awsm ec2 start i-1234567890abcdef0
```

#### Stop an EC2 Instance

```bash
awsm ec2 stop <instance-id>
```

Example:
```bash
awsm ec2 stop i-1234567890abcdef0
```

### S3 Commands

#### List S3 Buckets

```bash
awsm s3 ls
```

#### List Objects in a Bucket

```bash
awsm s3 ls <bucket-name> [--prefix <prefix>] [--max-items <number>]
```

Example:
```bash
# List all objects in a bucket
awsm s3 ls my-bucket

# List objects with a specific prefix
awsm s3 ls my-bucket --prefix "logs/"

# Limit the number of objects returned
awsm s3 ls my-bucket --max-items 100
```

#### Upload a File to S3

```bash
awsm s3 cp <local-file> s3://<bucket-name>/<key>
```

Example:
```bash
awsm s3 cp local-file.txt s3://my-bucket/remote-file.txt
```

#### Download a File from S3

```bash
awsm s3 cp s3://<bucket-name>/<key> <local-file>
```

Example:
```bash
awsm s3 cp s3://my-bucket/remote-file.txt local-file.txt
```

#### Delete an Object from S3

```bash
awsm s3 rm s3://<bucket-name>/<key>
```

Example:
```bash
awsm s3 rm s3://my-bucket/remote-file.txt
```

### Lambda Commands

#### List Lambda Functions

```bash
awsm lambda list [--max-items <number>]
```

Example:
```bash
# List all functions
awsm lambda list

# Limit the number of functions returned
awsm lambda list --max-items 10
```

#### Describe a Lambda Function

```bash
awsm lambda describe <function-name>
```

Example:
```bash
awsm lambda describe my-function
```

#### Invoke a Lambda Function

```bash
awsm lambda invoke <function-name> [--payload <json-string>] [--output-file <file>]
```

Example:
```bash
# Invoke with a JSON payload
awsm lambda invoke my-function --payload '{"key": "value"}'

# Save the response to a file
awsm lambda invoke my-function --payload '{"key": "value"}' --output-file response.json
```

#### View Lambda Function Logs

```bash
awsm lambda logs <function-name> [--start-time <time>] [--limit <number>]
```

Example:
```bash
# View recent logs
awsm lambda logs my-function

# View logs from a specific time
awsm lambda logs my-function --start-time "2023-01-01T00:00:00Z"

# Limit the number of log events
awsm lambda logs my-function --limit 100
```

## Terminal User Interface (TUI)

AWSM provides a terminal user interface (TUI) for managing AWS resources. To launch the TUI:

```bash
awsm tui
```

### Navigation

- Use arrow keys to navigate
- Press `Tab` to switch between panels
- Press `?` to show help
- Press `Esc` to go back or close dialogs
- Press `Ctrl+C` to exit

### Dashboard

The dashboard provides an overview of your AWS resources:

- EC2 instances (running, stopped, total)
- S3 buckets and objects
- Lambda functions

### EC2 View

The EC2 view allows you to manage EC2 instances:

- View instance details
- Start instances
- Stop instances
- Filter instances by state, type, or tags

### S3 View

The S3 view allows you to manage S3 buckets and objects:

- Browse buckets
- Browse objects within buckets
- Upload files
- Download files
- Delete objects

### Lambda View

The Lambda view allows you to manage Lambda functions:

- View function details
- Invoke functions
- View function logs

### Command Palette

Press `Ctrl+P` to open the command palette, which provides quick access to commands:

- Switch contexts
- Switch views
- Execute commands

### Context Switching

Press `Ctrl+X` to open the context switcher, which allows you to switch between contexts.

## Output Formatting

AWSM supports multiple output formats:

### Text Format (Default)

```bash
awsm ec2 list --output text
```

### JSON Format

```bash
awsm ec2 list --output json
```

### YAML Format

```bash
awsm ec2 list --output yaml
```

## Environment Variables

AWSM respects the following environment variables:

- `AWS_PROFILE`: AWS profile to use
- `AWS_REGION`: AWS region to use
- `AWS_ACCESS_KEY_ID`: AWS access key ID
- `AWS_SECRET_ACCESS_KEY`: AWS secret access key
- `AWS_SESSION_TOKEN`: AWS session token
- `AWSM_CONFIG_FILE`: Path to the AWSM configuration file
- `AWSM_OUTPUT_FORMAT`: Output format (text, json, yaml)

## Configuration File

AWSM uses a configuration file located at `~/.awsm.yaml`. The file has the following structure:

```yaml
aws:
  profile: default
  region: us-west-2
  role: ""
output:
  format: text
contexts:
  default:
    profile: default
    region: us-west-2
    role: ""
  dev:
    profile: development
    region: us-west-2
    role: ""
  prod:
    profile: production
    region: us-east-1
    role: "arn:aws:iam::123456789012:role/admin"
current_context: default
```

## Advanced Usage

### Using AWS IAM Roles

To use an IAM role:

```bash
awsm context create role-context --profile base-profile --region us-west-2 --role arn:aws:iam::123456789012:role/role-name
awsm context use role-context
```

### Using AWS SSO

If you have AWS SSO configured in your AWS CLI, you can use it with AWSM:

```bash
aws sso login --profile sso-profile
awsm context create sso-context --profile sso-profile --region us-west-2
awsm context use sso-context
```

### Scripting with AWSM

AWSM can be used in scripts by using the JSON or YAML output format and parsing the output:

```bash
# Get a list of running EC2 instances as JSON
instances=$(awsm ec2 list --filter "instance-state-name=running" --output json)

# Parse the JSON with jq
instance_ids=$(echo "$instances" | jq -r '.[].ID')

# Iterate over the instance IDs
for id in $instance_ids; do
  echo "Processing instance $id"
  # Do something with the instance
done
```

### Using AWSM with AWS CloudShell

AWSM can be used with AWS CloudShell:

1. Download the Linux binary for AWSM
2. Upload it to CloudShell
3. Make it executable: `chmod +x awsm`
4. Run AWSM: `./awsm`

### Using AWSM with AWS CloudFormation

AWSM can be used to manage resources created by AWS CloudFormation:

```bash
# List EC2 instances created by a specific CloudFormation stack
awsm ec2 list --filter "tag:aws:cloudformation:stack-name=my-stack"
```

### Using AWSM with Docker

AWSM can be used with Docker:

```bash
docker run -it --rm \
  -v ~/.aws:/root/.aws \
  -v ~/.awsm.yaml:/root/.awsm.yaml \
  aoaws/awsm ec2 list