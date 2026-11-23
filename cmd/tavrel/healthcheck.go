package main

import (
	"fmt"
	"net/http"
	"time"
)

func healthcheck(addr string) error {
	client := http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get("http://" + addr + "/healthz")
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	return nil
}
