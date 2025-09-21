# BogdanFinn Go Wrapper

A powerful and user-friendly Go wrapper for the [bogdanfinn/tls-client](https://github.com/bogdanfinn/tls-client) library, providing advanced HTTP client capabilities with TLS fingerprinting, session management, and retry logic.

## 🚀 Features

- **Advanced TLS Fingerprinting**: Mimics real browser TLS signatures
- **Session Management**: Persistent cookies and headers across requests
- **Multiple HTTP Methods**: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- **Automatic Retries**: Configurable retry logic with exponential backoff
- **Body Format Support**: JSON, Form URL-encoded, Multipart, Plain text
- **Cookie Management**: Easy cookie manipulation and retrieval
- **Proxy Support**: HTTP/HTTPS/SOCKS proxy configuration
- **Type Safety**: Strong typing with proper error handling

## 📦 Installation

```bash
go get github.com/your-username/bogdanfinn-go-wrapper
```

## 🔧 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    wrapper "github.com/your-username/bogdanfinn-go-wrapper"
)

func main() {
    // Create a new session
    session, err := wrapper.NewSession(nil)
    if err != nil {
        log.Fatal("Failed to create session:", err)
    }
    defer session.Close()

    // Simple GET request
    response := session.Get(wrapper.RequestOptions{
        Url: "https://httpbin.org/get",
        Headers: map[string]string{
            "User-Agent": "MyApp/1.0",
        },
    })

    if response.Error != "" {
        log.Fatal("Request failed:", response.Error)
    }

    fmt.Printf("Status: %d\n", response.StatusCode)
    fmt.Printf("Body: %s\n", string(response.Body))
}
```

### Advanced Configuration

```go
config := &wrapper.SessionConfig{
    UserAgent:       stringPtr("CustomBot/2.0"),
    Timeout:         intPtr(30),
    MaxRetries:      intPtr(5),
}

session, err := wrapper.NewSession(config)
if err != nil {
    log.Fatal(err)
}
```

## 📋 Examples

### POST Request with JSON Body

```go
// JSON POST request
response := session.Post(wrapper.RequestOptions{
    Url: "https://httpbin.org/post",
    Headers: map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer your-token",
    },
    Body: map[string]interface{}{
        "username": "john_doe",
        "email":    "john@example.com",
        "age":      30,
    },
})

fmt.Printf("Response: %s\n", string(response.Body))
```

### Form Data POST

```go
// Form URL-encoded POST
response := session.Post(wrapper.RequestOptions{
    Url: "https://httpbin.org/post",
    Headers: map[string]string{
        "Content-Type": "application/x-www-form-urlencoded",
    },
    Body: map[string]string{
        "username": "john_doe",
        "password": "secret123",
    },
})
```

### Multipart Form Data

```go
// Multipart form data
response := session.Post(wrapper.RequestOptions{
    Url: "https://httpbin.org/post",
    Headers: map[string]string{
        "Content-Type": "multipart/form-data",
    },
    Body: map[string]string{
        "file_field": "file_content_here",
        "text_field": "some text value",
    },
})
```

### URL Parameters

```go
// GET with query parameters
response := session.Get(wrapper.RequestOptions{
    Url: "https://httpbin.org/get",
    Params: map[string]string{
        "page":  "1",
        "limit": "10",
        "sort":  "name",
    },
})
// Final URL: https://httpbin.org/get?page=1&limit=10&sort=name
```

### Cookie Management

```go
// Set a cookie
targetURL, _ := url.Parse("https://example.com")
err := session.SetCookies("session_id", "abc123", targetURL)
if err != nil {
    log.Fatal(err)
}

// Get a specific cookie
if cookieValue := session.GetCookie("session_id", targetURL); cookieValue != nil {
    fmt.Printf("Session ID: %s\n", *cookieValue)
} else {
    fmt.Println("Cookie not found")
}

// Get all cookies
cookies, err := session.GetCookies(targetURL)
if err != nil {
    log.Fatal(err)
}

for name, value := range cookies {
    fmt.Printf("Cookie %s = %s\n", name, value)
}

// Clear all cookies
err = session.ClearCookies()
if err != nil {
    log.Fatal(err)
}
```

### Custom Cookie with Options

```go
customCookie := &http.Cookie{
    Name:     "custom_cookie",
    Value:    "custom_value",
    MaxAge:   3600,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode,
    Domain:   ".example.com",
    Path:     "/api",
}

err := session.SetCookiesWithOptions(customCookie, targetURL)
if err != nil {
    log.Fatal(err)
}
```

### Proxy Configuration

```go
// Set proxy
err := session.SetProxy("http://proxy.example.com:8080")
if err != nil {
    log.Fatal("Failed to set proxy:", err)
}

// Or with authentication
err = session.SetProxy("http://username:password@proxy.example.com:8080")
if err != nil {
    log.Fatal("Failed to set proxy:", err)
}
```

### Error Handling and Retries

```go
response := session.Get(wrapper.RequestOptions{
    Url: "https://unreliable-api.com/data",
    Headers: map[string]string{
        "Accept": "application/json",
    },
})

// Check for errors
if response.Error != "" {
    fmt.Printf("Request failed after retries: %s\n", response.Error)
    return
}

// Check HTTP status
if response.StatusCode >= 400 {
    fmt.Printf("HTTP Error: %d\n", response.StatusCode)
    return
}

// Process successful response
fmt.Printf("Success: %s\n", string(response.Body))
```

### Session Information and Validation

```go
// Check if session is valid
if !session.IsValid() {
    log.Fatal("Session is no longer valid")
}

// Get session information
info := session.GetSessionInfo()
fmt.Printf("Session Info: %+v\n", info)
```

### Complete Example: Web Scraping with Login

```go
package main

import (
    "fmt"
    "log"
    "net/url"
    wrapper "github.com/your-username/bogdanfinn-go-wrapper"
)

func main() {
    // Create session
    config := &wrapper.SessionConfig{
        UserAgent:  stringPtr("Mozilla/5.0 (compatible; WebScraper/1.0)"),
        MaxRetries: intPtr(3),
        Timeout:    intPtr(30),
    }

    session, err := wrapper.NewSession(config)
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    baseURL := "https://example.com"
    loginURL := baseURL + "/login"
    targetURL, _ := url.Parse(baseURL)

    // Step 1: Login
    loginResponse := session.Post(wrapper.RequestOptions{
        Url: loginURL,
        Headers: map[string]string{
            "Content-Type": "application/x-www-form-urlencoded",
            "Referer":      baseURL,
        },
        Body: map[string]string{
            "username": "your_username",
            "password": "your_password",
        },
    })

    if loginResponse.Error != "" || loginResponse.StatusCode != 200 {
        log.Fatal("Login failed")
    }

    // Step 2: Check if we got authentication cookie
    authCookie := session.GetCookie("auth_token", targetURL)
    if authCookie == nil {
        log.Fatal("No authentication cookie received")
    }

    fmt.Printf("Logged in successfully, auth token: %s\n", *authCookie)

    // Step 3: Access protected content
    protectedResponse := session.Get(wrapper.RequestOptions{
        Url: baseURL + "/protected-data",
        Headers: map[string]string{
            "Accept": "application/json",
        },
    })

    if protectedResponse.Error != "" {
        log.Fatal("Failed to access protected content:", protectedResponse.Error)
    }

    fmt.Printf("Protected data: %s\n", string(protectedResponse.Body))
}

// Helper functions
func stringPtr(s string) *string { return &s }
func intPtr(i int) *int         { return &i }
```

## 🔧 Configuration Options

### SessionConfig

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `UserAgent` | `*string` | Chrome UA | Custom User-Agent string |
| `SecChUa` | `*string` | Chrome value | SEC-CH-UA header |
| `SecChUaPlatform` | `*string` | Platform value | SEC-CH-UA-Platform header |
| `Timeout` | `*int` | 60 | Request timeout in seconds |
| `Profile` | `*tls_client.ClientProfile` | Chrome_133 | TLS client profile |
| `MaxRetries` | `*int` | 3 | Maximum retry attempts |

### RequestOptions

| Field | Type | Description |
|-------|------|-------------|
| `Url` | `string` | Target URL (required) |
| `Headers` | `map[string]string` | HTTP headers |
| `Body` | `interface{}` | Request body (various types supported) |
| `Params` | `map[string]string` | URL query parameters |

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [bogdanfinn/tls-client](https://github.com/bogdanfinn/tls-client) - The underlying TLS client library
- [bogdanfinn/fhttp](https://github.com/bogdanfinn/fhttp) - HTTP library with fingerprinting support

## 📞 Support

If you have any questions or issues, please open an issue on the GitHub repository.