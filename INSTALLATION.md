# AWSM Installation Guide

This guide provides detailed instructions for installing AWSM on different platforms.

## Prerequisites

Before installing AWSM, ensure you have the following prerequisites:

- **Go 1.18 or later** - Required for building from source
- **AWS credentials** - Configured via `~/.aws/credentials` or environment variables
- **Git** - Required for cloning the repository (if building from source)

## Installation Methods

### 1. Using Pre-built Binaries (Recommended)

#### Linux and macOS

```bash
# Download the latest release for your platform
curl -L https://github.com/ao/awsm/releases/latest/download/awsm-$(uname -s)-$(uname -m) -o awsm

# Make the binary executable
chmod +x awsm

# Move the binary to a directory in your PATH
sudo mv awsm /usr/local/bin/
```

#### Windows

1. Download the latest Windows release from the [Releases page](https://github.com/ao/awsm/releases)
2. Rename the downloaded file to `awsm.exe`
3. Move the file to a directory in your PATH (e.g., `C:\Windows\System32\`)

### 2. Building from Source

#### Prerequisites

- Go 1.18 or later
- Git

#### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/ao/awsm.git
   cd awsm
   ```

2. Build the binary:
   ```bash
   go build -o awsm ./cmd/awsm
   ```

3. Install the binary:
   ```bash
   # Linux/macOS
   sudo mv awsm /usr/local/bin/

   # Windows (run in Command Prompt as Administrator)
   move awsm.exe C:\Windows\System32\
   ```

### 3. Using Go Install

If you have Go installed, you can use the `go install` command:

```bash
go install github.com/ao/awsm/cmd/awsm@latest
```

This will install the `awsm` binary in your `$GOPATH/bin` directory. Make sure this directory is in your PATH.

### 4. Using Package Managers

#### Homebrew (macOS and Linux)

```bash
brew tap ao/awsm
brew install awsm
```

#### Scoop (Windows)

```powershell
scoop bucket add ao https://github.com/ao/scoop-bucket.git
scoop install awsm
```

## Verifying the Installation

After installation, verify that AWSM is installed correctly:

```bash
awsm --version
```

You should see output similar to:

```
awsm version 1.0.0
```

## Initial Configuration

After installing AWSM, you should initialize the configuration:

```bash
awsm config init
```

This will create a configuration file at `~/.awsm.yaml` with default settings.

## AWS Credentials

AWSM uses the same credential sources as the AWS CLI:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, etc.)
2. AWS credentials file (`~/.aws/credentials`)
3. IAM roles for Amazon EC2 or ECS tasks

If you haven't configured AWS credentials yet, you can do so using:

```bash
aws configure
```

Or manually create the credentials file:

```bash
mkdir -p ~/.aws
cat > ~/.aws/credentials << EOF
[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
EOF
```

## Troubleshooting

### Common Issues

#### Permission Denied

If you encounter a "permission denied" error when running AWSM:

```bash
chmod +x /path/to/awsm
```

#### Command Not Found

If you encounter a "command not found" error:

1. Ensure the binary is in a directory in your PATH
2. Check your PATH environment variable:
   ```bash
   echo $PATH
   ```

#### AWS Credential Issues

If you encounter AWS credential issues:

1. Verify your credentials are correctly configured:
   ```bash
   aws configure list
   ```
2. Check the AWS credentials file:
   ```bash
   cat ~/.aws/credentials
   ```

## Upgrading

To upgrade AWSM to the latest version:

### Pre-built Binaries

Follow the same installation steps to download and install the latest version.

### From Source

```bash
cd /path/to/awsm/repository
git pull
go build -o awsm ./cmd/awsm
sudo mv awsm /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/ao/awsm/cmd/awsm@latest
```

### Using Package Managers

#### Homebrew

```bash
brew update
brew upgrade awsm
```

#### Scoop

```powershell
scoop update awsm
```

## Uninstallation

To uninstall AWSM:

### Manual Installation

```bash
# Linux/macOS
sudo rm /usr/local/bin/awsm

# Windows (run in Command Prompt as Administrator)
del C:\Windows\System32\awsm.exe
```

### Package Managers

#### Homebrew

```bash
brew uninstall awsm
```

#### Scoop

```powershell
scoop uninstall awsm
```

## Next Steps

After installing AWSM, refer to the [Usage Guide](USAGE.md) for information on how to use the tool.