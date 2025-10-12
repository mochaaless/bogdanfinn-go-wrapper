package bogdanfinn_go_wrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
)

// Format headers with default values and error handling
func formatHeaders(s *Session, headers map[string]string) (http.Header, error) {
	if headers == nil {
		return http.Header{}, nil
	}

	formattedHeaders := http.Header{}
	var order []string

	for key, value := range headers {
		// Validate that key is not empty
		if strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("header key cannot be empty")
		}

		normalizedKey := strings.ToLower(key)

		// Apply default values
		switch normalizedKey {
		case "user-agent":
			if value == "" || value == "default" {
				value = s.UserAgent
			}
		case "sec-ch-ua":
			if value == "" || value == "default" {
				value = s.SecChUa
			}
		case "sec-ch-ua-platform":
			if value == "" || value == "default" {
				value = s.SecChUaPlatform
			}
		}

		// Use original key (with correct case)
		formattedHeaders[key] = []string{value}
		order = append(order, key)
	}

	formattedHeaders[http.HeaderOrderKey] = order
	return formattedHeaders, nil
}

// Format body with improved error handling and content-type detection
func formatBody(headers map[string]string, body interface{}) (io.Reader, string, error) {
	if body == nil {
		return nil, "", nil
	}

	// Look for content-type (case insensitive)
	var contentType string
	for key, value := range headers {
		if strings.ToLower(key) == "content-type" {
			contentType = strings.ToLower(strings.TrimSpace(value))
			break
		}
	}

	// If no content-type specified, infer from body type
	if contentType == "" {
		switch body.(type) {
		case string:
			contentType = "text/plain"
		default:
			contentType = "application/json"
		}
	}

	switch {
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return formatFormURLEncoded(body)

	case strings.Contains(contentType, "text/plain"):
		return formatTextPlain(body)

	case strings.Contains(contentType, "multipart/form-data"):
		reader, newContentType, err := formatMultipartFormData(body)
		return reader, newContentType, err

	default: // JSON by default
		return formatJSON(body)
	}
}

func formatFormURLEncoded(body interface{}) (io.Reader, string, error) {
	formData, ok := body.(map[string]string)
	if !ok {
		return nil, "", fmt.Errorf("body must be map[string]string for application/x-www-form-urlencoded")
	}

	data := url.Values{}
	for k, v := range formData {
		data.Set(k, v)
	}

	return strings.NewReader(data.Encode()), "application/x-www-form-urlencoded", nil
}

func formatTextPlain(body interface{}) (io.Reader, string, error) {
	bodyStr, ok := body.(string)
	if !ok {
		return nil, "", fmt.Errorf("body must be string for text/plain")
	}
	return strings.NewReader(bodyStr), "text/plain", nil
}

func formatMultipartFormData(body interface{}) (io.Reader, string, error) {
	formData, ok := body.(map[string]string)
	if !ok {
		return nil, "", fmt.Errorf("body must be map[string]string for multipart/form-data")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for k, v := range formData {
		if err := writer.WriteField(k, v); err != nil {
			writer.Close() // Close writer on error
			return nil, "", fmt.Errorf("error writing field %s: %w", k, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("error closing multipart writer: %w", err)
	}

	return &buf, writer.FormDataContentType(), nil
}

func formatJSON(body interface{}) (io.Reader, string, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("error marshaling body to JSON: %w", err)
	}
	return strings.NewReader(string(bodyJSON)), "application/json", nil
}

// Build URL with params
func buildURL(baseURL string, params map[string]string) (string, error) {
	if strings.TrimSpace(baseURL) == "" {
		return "", fmt.Errorf("baseURL cannot be empty")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	if params == nil || len(params) == 0 {
		return parsedURL.String(), nil
	}

	queryParams := parsedURL.Query() // Preserve existing parameters

	for key, value := range params {
		if strings.TrimSpace(key) == "" {
			continue // Ignore empty keys instead of failing
		}
		queryParams.Add(key, value)
	}

	parsedURL.RawQuery = queryParams.Encode()
	return parsedURL.String(), nil
}

// Response handler with detailed error reporting
func handleResponse(s *Session, request RequestOptions, req *http.Request, err error) Response {
	emptyResponse := Response{
		Url:        nil,
		Cookies:    nil,
		Headers:    nil,
		Body:       nil,
		StatusCode: 0,
	}

	if err != nil {
		emptyResponse.Error = fmt.Sprintf("Error creating request: %v", err)
		return emptyResponse
	}

	if !s.IsValid() {
		emptyResponse.Error = "Session or client is nil"
		return emptyResponse
	}

	// Format headers with error handling
	headers, err := formatHeaders(s, request.Headers)
	if err != nil {
		emptyResponse.Error = fmt.Sprintf("Error formatting headers: %v", err)
		return emptyResponse
	}
	req.Header = headers

	resp, err := s.Client.Do(req)
	if err != nil {
		emptyResponse.Error = fmt.Sprintf("Request error: %v", err)
		return emptyResponse
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log close error if you have logging
			_ = closeErr
		}
	}()

	// Get response URL
	respURL := req.URL
	if location, err := resp.Location(); err == nil && location != nil {
		respURL = location
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		emptyResponse.Error = fmt.Sprintf("Error reading response body: %v", err)
		return emptyResponse
	}

	return Response{
		Url:        respURL,
		Cookies:    resp.Cookies(),
		Headers:    resp.Header,
		Body:       body,
		StatusCode: resp.StatusCode,
		Error:      "",
	}
}

// Utility functions
func getIntOrDefault(value *int, defaultValue int) int {
	if value != nil {
		return *value
	}
	return defaultValue
}

func getStringOrDefault(value *string, defaultValue string) string {
	if value != nil && *value != "" {
		return *value
	}
	return defaultValue
}

// ValidateURL checks if a URL is valid and accessible
func ValidateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http/https)")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include host")
	}

	return nil
}

// SanitizeHeaders removes empty or invalid headers
func SanitizeHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return make(map[string]string)
	}

	sanitized := make(map[string]string)
	for key, value := range headers {
		if strings.TrimSpace(key) != "" {
			sanitized[key] = value
		}
	}

	return sanitized
}
