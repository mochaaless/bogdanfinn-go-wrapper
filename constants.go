package bogdanfinn_go_wrapper


// Default values for headers
var user_agent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"
var sech_ua string = `"Not;A=Brand";v="99", "Google Chrome";v="139", "Chromium";v="139"`
var sech_ua_platform string = `"macOS"`


// Common session errors for retry logic
var sessionErrors []string = []string{
	"TLS handshake timeout",
	"Proxy responded with non 200 code",
	"no such host",
	"EOF",
	"410 Gone",
	"407 Proxy Authentication Required",
	"417 Expectation Failed",
	"unknown authority",
	"certificate: x509",
	"while awaiting headers",
	"remote error",
}
