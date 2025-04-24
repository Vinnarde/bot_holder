# Redirector App

A Go Fiber application that handles redirection based on bot validation parameters for multiple domains.

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

The application uses a YAML configuration file (`config.yaml`) that supports multiple domains. Each domain can have its own settings:

```yaml
port: 8080                                     # Server port
domains:
  cskinmasters.com:                            # Domain name
    base_redirect_url: "https://cskinmasters.com"  # URL to redirect to when validation fails
    expect_bot_param: "request_hash"               # Query parameter for bot validation
    expect_bot_value: "v07N4wzsJHP5mCRn44x6Mtzib8VeFzTC"  # Expected value for the bot parameter
    bot_cookie_name: "visit"                       # Cookie name for validation
    bot_cookie_value: "P5mCRn44x6Mtzib"            # Cookie value for validation
    page_template: "index%d.html"                  # Template for page URL generation
    min_redirect_seconds: 3                        # Minimum seconds before redirect
    max_redirect_seconds: 5                        # Maximum seconds before redirect
  example.com:                                # Another domain with different settings
    base_redirect_url: "https://example.com"
    expect_bot_param: "request_hash"
    expect_bot_value: "different_value_for_example"
    bot_cookie_name: "visit"
    bot_cookie_value: "different_cookie_value"
    page_template: "page%d.html"
    min_redirect_seconds: 2
    max_redirect_seconds: 4
```

### Live Configuration Reload

The application supports live configuration reloading. Any changes made to `config.yaml` will be automatically detected and applied without restarting the application. This allows for dynamic updates to:

- Adding new domains
- Modifying existing domain settings
- Redirect timings
- Bot validation parameters
- Server port (will apply on next server start)
- Redirect URLs

## Running the application

```
go run main.go
```

The server will start on the port specified in the configuration file.

## Domain Handling

The application automatically detects the domain from the request's host header and applies the corresponding configuration. It supports:

- Exact domain matches (e.g., example.com)
- Subdomain matches (e.g., sub.example.com will use example.com's config)
- Localhost and IP addresses

## API Endpoints

- `/` - Main entry point
- `/index{id}.html` - Dynamic pages with IDs
- `/status` - Returns current configuration and server status for the domain

## How it works

1. Detects the domain from the request
2. Loads domain-specific configuration
3. Validates visitors using either the query parameter or the cookie defined in the config
4. If validation fails, redirects to the domain's base URL
5. If validation passes, sets a cookie and renders the page
6. The page will automatically redirect to a randomly numbered page after the configured time

## Routes

- `/` - Main entry point
- `/index{id}.html` - Dynamic pages with IDs 