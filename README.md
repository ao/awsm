# AWSM - AWS CLI Made Awesome

AWSM (pronounced "awesome") is a modern, user-friendly command-line interface for managing AWS resources. It provides a more intuitive and interactive experience compared to the standard AWS CLI, with features like context switching, a terminal UI mode, and simplified commands.

## Features

- **Terminal User Interface (TUI)** - Interactive terminal UI for managing AWS resources
- **Context Management** - Easily switch between different AWS profiles, regions, and roles
- **Simplified Commands** - More intuitive command structure compared to the standard AWS CLI
- **EC2 Management** - List, start, stop, and describe EC2 instances
- **S3 Management** - List buckets, list objects, upload, download, and delete objects
- **Lambda Management** - List, invoke, and view logs for Lambda functions
- **Output Formatting** - Format output as text, JSON, or YAML

## Installation

### Prerequisites

- Go 1.18 or later
- AWS credentials configured (via `~/.aws/credentials` or environment variables)

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/ao/awsm.git
   cd awsm
   ```

2. Build the binary:
   ```bash
   make build
   ```

3. Install the binary:
   ```bash
   make install
   ```

## Quick Start

### Configuration

AWSM uses a configuration file located at `~/.awsm.yaml`. You can initialize it with:

```bash
awsm config init
```

### Context Management

Create a new context:

```bash
awsm context create my-dev --profile dev --region us-west-2
```

Switch to a context:

```bash
awsm context use my-dev
```

List available contexts:

```bash
awsm context list
```

### EC2 Commands

List EC2 instances:

```bash
awsm ec2 list
```

Start an EC2 instance:

```bash
awsm ec2 start i-1234567890abcdef0
```

Stop an EC2 instance:

```bash
awsm ec2 stop i-1234567890abcdef0
```

### S3 Commands

List S3 buckets:

```bash
awsm s3 ls
```

List objects in a bucket:

```bash
awsm s3 ls my-bucket
```

Upload a file to S3:

```bash
awsm s3 cp local-file.txt s3://my-bucket/remote-file.txt
```

Download a file from S3:

```bash
awsm s3 cp s3://my-bucket/remote-file.txt local-file.txt
```

### Lambda Commands

List Lambda functions:

```bash
awsm lambda list
```

Invoke a Lambda function:

```bash
awsm lambda invoke my-function --payload '{"key": "value"}'
```

View Lambda function logs:

```bash
awsm lambda logs my-function
```

### Terminal UI Mode

Launch the terminal UI:

```bash
awsm tui
```

In TUI mode, you can:
- Navigate between different AWS services
- View and manage EC2 instances
- Browse S3 buckets and objects
- View and invoke Lambda functions
- Switch between contexts

## Documentation

For more detailed documentation, see:

- [Installation Guide](INSTALLATION.md)
- [Usage Guide](USAGE.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)

## Development

### CI/CD Pipeline

This project uses GitHub Actions for continuous integration and deployment:

- **CI Workflow**: Runs on pull requests and pushes to main
  - Runs unit and integration tests
  - Performs code linting
  - Reports code coverage
  - Tests on multiple Go versions and operating systems

- **Release Workflow**: Runs when a new tag is pushed
  - Builds binaries for multiple platforms (Linux, macOS, Windows)
  - Creates GitHub releases
  - Attaches binaries to releases
  - Generates release notes
  - Publishes to GitHub Packages

### Release Process

To create a new release:

1. Update the version in `version.go`
2. Update the `CHANGELOG.md` file
3. Run `make bump-version VERSION=x.y.z`
4. Run `make release-create`
5. Push the new tag: `git push origin vx.y.z`

The release workflow will automatically build and publish the release.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.