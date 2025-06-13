package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FormatPatchVersion removes trailing ".0" from patch versions if present.
// Example: "15.07" -> "15.7", "15.10" -> "15.10"
func FormatPatchVersion(patch string) string {
	split := strings.Split(patch, ".")
	if len(split) != 2 {
		// Return original string if format is unexpected
		return patch
	}
	minor := split[1]
	if len(minor) == 2 && strings.HasPrefix(minor, "0") {
		minor = strings.TrimPrefix(minor, "0")
	}
	return split[0] + "." + minor
}

// op.gg is expecting the patch version in the format of "15.07"
// This function will take a patch version in the format of "15.7" and return "15.07". but 15.11 will return 15.11
func FormatPatchVersionForOpGG(patch string) string {
	split := strings.Split(patch, ".")
	if len(split[1]) == 1 {
		return split[0] + ".0" + split[1]
	}
	return patch
}

// GetLatestPatchVersion fetches the list of patches from Riot's API and returns the latest one
// in Major.Minor format (e.g., "15.8").
func GetLatestPatchVersion(apiUrl string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch patch versions from %s: %w", apiUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch patch versions: received status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var versions []string
	err = json.Unmarshal(body, &versions)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal patch versions JSON: %w", err)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no patch versions found in the API response")
	}

	// The first version in the list is the latest one (e.g., "15.8.1")
	latestVersion := versions[0]

	// Extract Major.Minor (e.g., "15.8" from "15.8.1")
	split := strings.Split(latestVersion, ".")
	if len(split) < 2 {
		return "", fmt.Errorf("unexpected version format received: %s", latestVersion)
	}

	formattedPatch := split[0] + "." + split[1]
	return formattedPatch, nil
}
