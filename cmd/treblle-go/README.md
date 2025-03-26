# Treblle Go SDK CLI

The Treblle Go SDK CLI provides a command-line interface for debugging and inspecting the Treblle SDK configuration.

## Usage

### Running without installation

```bash
go run github.com/treblle/treblle-go/cmd/treblle-go -debug
```

### Installing and using the CLI

```bash
go install github.com/treblle/treblle-go/cmd/treblle-go@latest
treblle-go -debug
```

## Available Commands

- `-debug`: Shows SDK configuration information

## Debug Output

The debug command displays:

1. SDK Version
2. Project ID (masked for security)
3. API Key (masked for security)
4. Configured Treblle URL
5. Ignored Environments

This is particularly useful when troubleshooting issues with the Treblle SDK or verifying that your configuration is correct.

## Environment Variables

The CLI tool respects the following environment variables:

- `TREBLLE_API_KEY`: Your Treblle API key
- `TREBLLE_PROJECT_ID`: Your Treblle project ID
- `TREBLLE_ENDPOINT`: Custom Treblle API endpoint (optional)
- `TREBLLE_IGNORED_ENVIRONMENTS`: Comma-separated list of environments to ignore (optional)
