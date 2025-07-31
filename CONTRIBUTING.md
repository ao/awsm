# Contributing to AWSM

Thank you for your interest in contributing to AWSM! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Contributing to AWSM](#contributing-to-awsm)
  - [Table of Contents](#table-of-contents)
  - [Getting Started](#getting-started)
  - [Development Environment](#development-environment)
    - [Prerequisites](#prerequisites)
    - [Setup](#setup)
  - [Project Structure](#project-structure)
  - [Coding Standards](#coding-standards)
    - [Go Code Style](#go-code-style)
    - [Commit Messages](#commit-messages)
  - [Testing](#testing)
    - [Running Tests](#running-tests)
    - [Writing Tests](#writing-tests)
    - [Test Organization](#test-organization)
  - [Pull Request Process](#pull-request-process)
  - [Documentation](#documentation)
    - [Code Documentation](#code-documentation)
    - [User Documentation](#user-documentation)
  - [Issue Reporting](#issue-reporting)
    - [Bug Reports](#bug-reports)
    - [Feature Requests](#feature-requests)
  - [Feature Requests](#feature-requests-1)
  - [Release Process](#release-process)
  - [License](#license)

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/awsm.git
   cd awsm
   ```
3. Add the upstream repository as a remote:
   ```bash
   git remote add upstream https://github.com/ao/awsm.git
   ```
4. Create a new branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Environment

### Prerequisites

- Go 1.18 or later
- Git
- Make
- AWS account (for testing)

### Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Build the project:
   ```bash
   make build
   ```

3. Run tests:
   ```bash
   make test
   ```

## Project Structure

The project is organized as follows:

```
awsm/
├── cmd/                  # Command-line interface
│   └── awsm/             # Main application entry point
├── internal/             # Internal packages
│   ├── aws/              # AWS service adapters
│   │   ├── ec2/          # EC2 adapter
│   │   ├── s3/           # S3 adapter
│   │   └── lambda/       # Lambda adapter
│   ├── config/           # Configuration management
│   ├── testutils/        # Testing utilities
│   └── tui/              # Terminal UI components
│       ├── components/   # Reusable UI components
│       └── models/       # UI models
├── tests/                # Integration tests
│   └── integration/      # Integration test files
├── docs/                 # Documentation
├── scripts/              # Build and utility scripts
└── examples/             # Example usage
```

## Coding Standards

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for issues
- Add comments for exported functions, types, and constants
- Write clear, concise, and descriptive comments
- Use meaningful variable and function names

### Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests after the first line
- Consider using the following format:
  ```
  [Component] Short description

  More detailed explanatory text, if necessary.

  Fixes #123
  ```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Run tests with coverage
make test-coverage
```

### Writing Tests

- Write unit tests for all new code
- Aim for at least 80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Use the `testutils` package for common testing utilities

### Test Organization

- Unit tests should be in the same package as the code they test
- Integration tests should be in the `tests/integration` directory
- Test files should be named `*_test.go`

## Pull Request Process

1. Ensure your code passes all tests and linting
2. Update the documentation with details of changes
3. Add or update tests as appropriate
4. Update the CHANGELOG.md with details of changes
5. Submit a pull request to the `main` branch
6. The PR will be reviewed by maintainers
7. Address any feedback from reviewers
8. Once approved, the PR will be merged by a maintainer

## Documentation

### Code Documentation

- Add comments for all exported functions, types, and constants
- Use [godoc](https://blog.golang.org/godoc-documenting-go-code) style comments
- Include examples where appropriate

### User Documentation

- Update README.md with any new features or changes
- Update USAGE.md with detailed usage instructions
- Update INSTALLATION.md if installation process changes

## Issue Reporting

### Bug Reports

When reporting a bug, please include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Screenshots or terminal output if applicable
- Environment information (OS, Go version, etc.)
- Any additional context

### Feature Requests

When requesting a feature, please include:

- A clear and descriptive title
- A detailed description of the feature
- The motivation for the feature
- Examples of how the feature would be used
- Any additional context

## Feature Requests

We welcome feature requests! Please submit them as issues with the "feature request" label.

When submitting a feature request, please:

1. Check if the feature has already been requested
2. Provide a clear and concise description of the feature
3. Explain why the feature would be useful
4. Provide examples of how the feature would be used
5. Consider the scope of the feature and how it fits with the project's goals

## Release Process

1. Update the version number in `cmd/awsm/main.go`
2. Update the CHANGELOG.md with the new version
3. Create a new tag with the version number:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   ```
4. Push the tag to GitHub:
   ```bash
   git push origin v1.0.0
   ```
5. GitHub Actions will automatically build and publish the release

## License

By contributing to AWSM, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).