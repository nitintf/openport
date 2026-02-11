package client

import "time"

// RequestLog captures metadata about a single proxied HTTP request.
type RequestLog struct {
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
	Timestamp  time.Time
}
