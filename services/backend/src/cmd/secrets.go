package main

import (
	"fmt"
	"os"
	"strings"
)

// secretVariables are the variables that can come from a file, as Docker
// Swarm hands secrets to a container: NAME_FILE names the file that holds the
// value of NAME.
var secretVariables = []string{
	"SIGTA_GOOGLE_CLIENT_ID",
	"SIGTA_GOOGLE_CLIENT_SECRET",
	"SIGTA_DATABASE_URL",
	"SIGTA_MIGRATIONS_DATABASE_URL",
}

// withSecretFiles reads the files of the secret variables once, at start, and
// returns a getenv that answers with their contents. A variable set both ways
// is an error, so it is never unclear which value is in use.
func withSecretFiles(getenv func(string) string, readFile func(string) ([]byte, error)) (func(string) string, error) {
	values := map[string]string{}
	for _, name := range secretVariables {
		path := getenv(name + "_FILE")
		if path == "" {
			continue
		}
		if getenv(name) != "" {
			return nil, fmt.Errorf("%s and %s_FILE are both set; set only one", name, name)
		}
		content, err := readFile(path)
		if err != nil {
			return nil, fmt.Errorf("%s_FILE: %w", name, err)
		}
		values[name] = strings.TrimRight(string(content), "\r\n")
	}
	return func(name string) string {
		if value, ok := values[name]; ok {
			return value
		}
		return getenv(name)
	}, nil
}

func environment() (func(string) string, error) {
	return withSecretFiles(os.Getenv, os.ReadFile)
}
