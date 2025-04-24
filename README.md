# Redirector App

A Go Fiber application that handles redirection based on bot validation parameters.

## Requirements

- Go 1.21 or higher
- Fiber framework

## Installation

1. Clone the repository
2. Install dependencies:
```
go mod download
```

## Configuration

The application uses a YAML configuration file (`config.yaml`) with the following parameters:

```yaml
port: 3000                                     # Server port
base_redirect_url: "https://cskinmasters.com"  # URL to redirect to when validation fails
expect_bot_param: "request_hash"               # Query parameter for bot validation
expect_bot_value: "v07N4wzsJHP5mCRn44x6Mtzib8VeFzTC"  # Expected value for the bot parameter
bot_cookie_name: "visit"                       # Cookie name for validation
bot_cookie_value: "P5mCRn44x6Mtzib"            # Cookie value for validation
page_template: "index%d.html"                  # Template for page URL generation
min_redirect_seconds: 25                       # Minimum seconds before redirect
max_redirect_seconds: 30                       # Maximum seconds before redirect
```

### Live Configuration Reload

The application supports live configuration reloading. Any changes made to `config.yaml` will be automatically detected and applied without restarting the application. This allows for dynamic updates to:

- Redirect timings
- Bot validation parameters
- Server port (will apply on next server start)
- Redirect URLs

## Running the application

```
go run main.go
```

The server will start on the port specified in the configuration file.

## API Endpoints

- `/` - Main entry point
- `/index{id}.html` - Dynamic pages with IDs
- `/status` - Returns current configuration and server status

## How it works

1. Validates visitors using either the query parameter or the cookie defined in the config
2. If validation fails, redirects to the base URL
3. If validation passes, sets a cookie and renders the page
4. The page will automatically redirect to a randomly numbered page after the configured time

## Routes

- `/` - Main entry point
- `/index{id}.html` - Dynamic pages with IDs 