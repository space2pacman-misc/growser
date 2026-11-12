package net

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetch скачивает URL и возвращает тело ответа (HTML).
func Fetch(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	// Многие сайты отдают 403 без User-Agent
	req.Header.Set("User-Agent", "MiniBrowser/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}