package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// probeURL is the catalog, which is public and reads the database.
const probeURL = "http://127.0.0.1:8080/procedure-types"

// probe asks the server in this container for the catalog, so a container
// whose server or database does not answer is unhealthy. The image has neither
// a shell nor curl: the healthcheck of the deployment runs "backend probe".
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
