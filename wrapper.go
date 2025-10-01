package bogdanfinn_go_wrapper

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// NewSession creates a new HTTP session with TLS client
func NewSession(config *SessionConfig) (*Session, error) {
	if config == nil {
		config = &SessionConfig{}
	}

	jar := tls_client.NewCookieJar()
	if jar == nil {
		return nil, fmt.Errorf("failed to create cookie jar")
	}

	profile := profiles.Chrome_133
	if config.Profile != nil {
		profile = *config.Profile
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(getIntOrDefault(config.Timeout, 60)),
		tls_client.WithClientProfile(profile),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("error creating TLS session: %w", err)
	}

	return &Session{
		Client:          client,
		UserAgent:       getStringOrDefault(config.UserAgent, user_agent),
		SecChUa:         getStringOrDefault(config.SecChUa, sech_ua),
		SecChUaPlatform: getStringOrDefault(config.SecChUaPlatform, sech_ua_platform),
		MaxRetries:      getIntOrDefault(config.MaxRetries, 3),
	}, nil
}

// NewSessionLegacy maintains backward compatibility with the old constructor
func NewSessionLegacy(ua, s_ua, s_ua_platform *string, timeout *int) (*Session, error) {
	config := &SessionConfig{
		UserAgent:       ua,
		SecChUa:         s_ua,
		SecChUaPlatform: s_ua_platform,
		Timeout:         timeout,
	}
	return NewSession(config)
}

func (s *Session) executeRequest(method string, request RequestOptions) Response {
	// Validate session
	if !s.IsValid() {
		return Response{
			Url:        nil,
			Cookies:    nil,
			Headers:    nil,
			Body:       nil,
			StatusCode: 0,
			Error:      "session or client is nil",
		}
	}

	// Validate and build URL
	if strings.TrimSpace(request.Url) == "" {
		return Response{
			Error: "URL cannot be empty",
		}
	}

	parsedUrl, err := buildURL(request.Url, request.Params)
	if err != nil {
		return Response{
			Error: fmt.Sprintf("error parsing URL: %w", err),
		}
	}

	// Format body for methods that support it
	var bodyReader, contentType, bodyErr = formatBody(request.Headers, request.Body)
	if bodyErr != nil {
		return Response{
			Error: fmt.Sprintf("error formatting body: %w", bodyErr),
		}
	}

	// Update content-type header if needed (for multipart)
	if contentType != "" && request.Headers != nil {
		if request.Headers == nil {
			request.Headers = make(map[string]string)
		}
		// Only update if multipart (which changes the boundary)
		if strings.Contains(contentType, "multipart/form-data") {
			request.Headers["content-type"] = contentType
		}
	}

	// Create request
	req, err := http.NewRequest(method, parsedUrl, bodyReader)
	if err != nil {
		return Response{
			Error: fmt.Sprintf("error creating request: %w", err),
		}
	}

	// Execute with retry logic
	return s.executeWithRetry(request, req)
}

// executeWithRetry handles the request execution with retry logic
func (s *Session) executeWithRetry(request RequestOptions, req *http.Request) Response {
	var lastResponse Response
	maxRetries := s.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1 // At least one attempt
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp := handleResponse(s, request, req, nil)

		// If no session error, return the response
		if resp.Error == "" {
			return resp
		}

		lastResponse = resp

		// Don't wait on the last attempt
		if attempt < maxRetries-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(attempt+1)) // Exponential backoff
		}
	}

	// If we get here, all retries failed
	lastResponse.Error = fmt.Sprintf("max retries (%d) exceeded: %s", maxRetries, lastResponse.Error)
	return lastResponse
}

// GET performs a GET request
func (s *Session) Get(request RequestOptions) Response {
	return s.executeRequest(http.MethodGet, request)
}

// POST performs a POST request
func (s *Session) Post(request RequestOptions) Response {
	return s.executeRequest(http.MethodPost, request)
}

// PUT performs a PUT request
func (s *Session) Put(request RequestOptions) Response {
	return s.executeRequest(http.MethodPut, request)
}

// DELETE performs a DELETE request
func (s *Session) Delete(request RequestOptions) Response {
	return s.executeRequest(http.MethodDelete, request)
}

// PATCH performs a PATCH request
func (s *Session) Patch(request RequestOptions) Response {
	return s.executeRequest(http.MethodPatch, request)
}

// HEAD performs a HEAD request
func (s *Session) Head(request RequestOptions) Response {
	return s.executeRequest(http.MethodHead, request)
}

// OPTIONS performs an OPTIONS request
func (s *Session) Options(request RequestOptions) Response {
	return s.executeRequest(http.MethodOptions, request)
}

// SetCookies sets a cookie for the session
func (s *Session) SetCookies(name, value string, targetURL *url.URL) error {
	if !s.IsValid() {
		return fmt.Errorf("session or client is nil")
	}

	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("cookie name cannot be empty")
	}

	if targetURL == nil {
		return fmt.Errorf("target URL cannot be nil")
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   60 * 60 * 24 * 365, // 1 year
		HttpOnly: true,
		Secure:   targetURL.Scheme == "https",
		SameSite: http.SameSiteNoneMode,
	}

	s.Client.SetCookies(targetURL, []*http.Cookie{cookie})
	return nil
}

// SetCookiesWithOptions sets a cookie with custom options
func (s *Session) SetCookiesWithOptions(cookie *http.Cookie, targetURL *url.URL) error {
	if !s.IsValid() {
		return fmt.Errorf("session or client is nil")
	}

	if cookie == nil {
		return fmt.Errorf("cookie cannot be nil")
	}

	if targetURL == nil {
		return fmt.Errorf("target URL cannot be nil")
	}

	if strings.TrimSpace(cookie.Name) == "" {
		return fmt.Errorf("cookie name cannot be empty")
	}

	s.Client.SetCookies(targetURL, []*http.Cookie{cookie})
	return nil
}

// SetProxy sets a proxy for the session
func (s *Session) SetProxy(proxy string) error {
	if !s.IsValid() {
		return fmt.Errorf("session or client is nil")
	}

	if strings.TrimSpace(proxy) == "" {
		return fmt.Errorf("proxy URL cannot be empty")
	}

	// Validate proxy URL format
	if _, err := url.Parse(proxy); err != nil {
		return fmt.Errorf("invalid proxy URL format: %w", err)
	}

	s.Client.SetProxy(proxy)
	return nil
}

// ClearCookies clears all cookies from the session
func (s *Session) ClearCookies() error {
	if !s.IsValid() {
		return fmt.Errorf("session or client is nil")
	}

	jar := tls_client.NewCookieJar()
	if jar == nil {
		return fmt.Errorf("failed to create new cookie jar")
	}

	s.Client.SetCookieJar(jar)
	return nil
}

// GetCookies returns all cookies with full details
func (s *Session) GetCookies(targetURL *url.URL) ([]*http.Cookie, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("session or client is nil")
	}

	cookies := s.Client.GetCookies(targetURL)
	validCookies := make([]*http.Cookie, 0, len(cookies))

	for _, cookie := range cookies {
		if cookie != nil && strings.TrimSpace(cookie.Name) != "" {
			validCookies = append(validCookies, cookie)
		}
	}

	return validCookies, nil
}

// GetCookie returns the value of a specific cookie by name, or nil if not found
func (s *Session) GetCookie(name string, targetURL *url.URL) *http.Cookie {
	if !s.IsValid() {
		return nil
	}

	if strings.TrimSpace(name) == "" {
		return nil
	}

	cookies := s.Client.GetCookies(targetURL)

	for _, cookie := range cookies {
		if cookie != nil && cookie.Name == name {
			return cookie
		}
	}

	return nil
}

// Close gracefully closes the session (if needed)
func (s *Session) Close() error {
	if !s.IsValid() {
		return nil // Already closed or nil
	}

	// Clear cookies as cleanup
	if err := s.ClearCookies(); err != nil {
		return fmt.Errorf("error clearing cookies during close: %w", err)
	}

	// Note: tls_client doesn't seem to have an explicit close method but we can nil the client reference
	s.Client = nil
	return nil
}

// IsValid checks if the session is still valid and usable
func (s *Session) IsValid() bool {
	return s != nil && s.Client != nil
}

// GetSessionInfo returns basic information about the session
func (s *Session) GetSessionInfo() map[string]interface{} {
	if !s.IsValid() {
		return map[string]interface{}{
			"valid": false,
		}
	}

	return map[string]interface{}{
		"valid":              true,
		"user_agent":         s.UserAgent,
		"sec_ch_ua":          s.SecChUa,
		"sec_ch_ua_platform": s.SecChUaPlatform,
		"max_retries":        s.MaxRetries,
	}
}
