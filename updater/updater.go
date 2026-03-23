// Package updater checks for new FRACTURE releases on GitHub.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	CurrentVersion = "1.5.0"
	repoOwner      = "mkvinicius"
	repoName       = "fracture"
	apiURL         = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

// ReleaseInfo holds information about the latest GitHub release.
type ReleaseInfo struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

// UpdateResult is returned by CheckForUpdate.
type UpdateResult struct {
	HasUpdate      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	ReleaseName    string
	ReleaseNotes   string
}

// CheckForUpdate queries the GitHub API and returns update info.
// It returns quickly (5s timeout) and never blocks startup.
func CheckForUpdate() (*UpdateResult, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "FRACTURE/"+CurrentVersion)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(CurrentVersion, "v")

	hasUpdate := compareVersions(latest, current) > 0

	return &UpdateResult{
		HasUpdate:      hasUpdate,
		CurrentVersion: current,
		LatestVersion:  latest,
		ReleaseURL:     release.HTMLURL,
		ReleaseName:    release.Name,
		ReleaseNotes:   release.Body,
	}, nil
}

// compareVersions returns 1 if a > b, -1 if a < b, 0 if equal.
// Handles semver like "1.3.0".
func compareVersions(a, b string) int {
	partsA := splitVersion(a)
	partsB := splitVersion(b)

	for i := 0; i < 3; i++ {
		va, vb := 0, 0
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &va)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &vb)
		}
		if va > vb {
			return 1
		}
		if va < vb {
			return -1
		}
	}
	return 0
}

func splitVersion(v string) []string {
	return strings.Split(v, ".")
}
