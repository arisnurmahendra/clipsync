package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type gasResp struct {
	OK   bool   `json:"ok"`
	Text string `json:"text"`
}

func fetchFromGAS(url string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Check if the response looks like JSON (starts with { or [)
	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	// Check if the response appears to be HTML instead of JSON
	if strings.HasPrefix(bodyStr, "<") {
		return "", fmt.Errorf("received HTML response instead of JSON: %.100s", bodyStr)
	}

	var result gasResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse json: %w. Response was: %.200s", err, bodyStr)
	}

	if !result.OK {
		return "", fmt.Errorf("API returned error: %s", result.Text)
	}

	return result.Text, nil
}