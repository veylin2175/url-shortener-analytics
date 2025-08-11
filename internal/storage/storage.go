package storage

import "errors"

var (
	ErrURLNotFound = errors.New("URL not found")
	ErrURLExists   = errors.New("URL already exists")
)

type AnalyticsData struct {
	TotalClicks int64
	UserAgents  map[string]int64
	Daily       map[string]int64
	Monthly     map[string]int64
}
