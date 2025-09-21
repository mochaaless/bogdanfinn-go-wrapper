package bogdanfinn_go_wrapper

import (
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"net/url"
)

// Session represents an HTTP session with TLS client
type Session struct {
	Client          tls_client.HttpClient
	UserAgent       string
	SecChUa         string
	SecChUaPlatform string
	MaxRetries      int
}

// SessionConfig holds configuration for creating a new session
type SessionConfig struct {
	UserAgent       *string
	SecChUa         *string
	SecChUaPlatform *string
	Timeout         *int
	Profile         *profiles.ClientProfile
	MaxRetries      *int
}

// Response represents an HTTP response
type Response struct {
	Url        *url.URL
	Body       []byte
	StatusCode int
	Headers    http.Header
	Cookies    []*http.Cookie
	Error      string
}

// RequestOptions holds options for making an HTTP request
type RequestOptions struct {
	Url     string
	Headers map[string]string
	Body    interface{}
	Params  map[string]string
}
