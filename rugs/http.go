package rugs

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	xml2json "github.com/basgys/goxml2json"
)

// HTTPGet performs a GET request with optional headers
func HTTPGet(timeout int, uri string, headers map[string]string) ([]byte, error) {
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// HTTPGetString performs a GET request and returns the response as a string
func HTTPGetString(timeout int, uri string, headers map[string]string) (string, error) {
	bytes, err := HTTPGet(timeout, uri, headers)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// HTTPPost performs a POST request with form values
func HTTPPost(timeout int, uri string, values map[string]string) ([]byte, error) {
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	formValues := url.Values{}
	for k, v := range values {
		formValues[k] = []string{v}
	}

	resp, err := client.PostForm(uri, formValues)
	if err != nil {
		return nil, fmt.Errorf("post failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// HTTPPostString performs a POST request and returns the response as a string
func HTTPPostString(timeout int, uri string, values map[string]string) (string, error) {
	bytes, err := HTTPPost(timeout, uri, values)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// XMLToJSON converts XML string to JSON string
func XMLToJSON(xml string) (string, error) {
	reader := strings.NewReader(xml)
	json, err := xml2json.Convert(reader)
	if err != nil {
		return "", fmt.Errorf("failed to convert XML to JSON: %w", err)
	}
	return json.String(), nil
}

// DetectContentType detects the MIME type of data
func DetectContentType(data []byte) string {
	return http.DetectContentType(data)
}

// ParseURL parses and validates a URL
func ParseURL(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	return u.String(), nil
}
