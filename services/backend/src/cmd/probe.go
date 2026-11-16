package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// The catalog is public and reads the database, so the probe also fails when
// the database does not answer.
const probeURL = "http://127.0.0.1:8080/procedure-types"

// probe is the container healthcheck: the image has no shell nor curl.
func probe(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", url, response.Status)
	}
	return nil
}
