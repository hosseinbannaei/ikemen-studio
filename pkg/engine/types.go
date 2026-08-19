package engine

import "time"

// ReleaseAsset represents an individual downloadable binary package
type ReleaseAsset struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"downloadUrl"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
}

// ReleaseInfo represents a GitHub release for Ikemen GO
type ReleaseInfo struct {
	Tag          string         `json:"tag"`
	Name         string         `json:"name"`
	PublishedAt  time.Time      `json:"publishedAt"`
	IsPrerelease bool           `json:"isPrerelease"`
	Body         string         `json:"body"`
	HTMLURL      string         `json:"htmlUrl"`
	Assets       []ReleaseAsset `json:"assets"`
}

// InstalledEngine represents a locally cached Ikemen GO version
type InstalledEngine struct {
	Version        string    `json:"version"`
	Path           string    `json:"path"`
	ExecutablePath string    `json:"executablePath"`
	InstalledAt    time.Time `json:"installedAt"`
	Channel        string    `json:"channel"`
	Size           int64     `json:"size"`
}

// DownloadProgress reports real-time progress for downloading and extracting
type DownloadProgress struct {
	Version         string  `json:"version"`
	Percent         float64 `json:"percent"`
	DownloadedBytes int64   `json:"downloadedBytes"`
	TotalBytes      int64   `json:"totalBytes"`
	Status          string  `json:"status"` // "downloading", "extracting", "completed", "error"
	Error           string  `json:"error,omitempty"`
}
