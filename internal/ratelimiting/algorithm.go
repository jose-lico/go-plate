package ratelimiting

import (
	"time"
)

type Algorithm interface {
	Allow(string) (bool, time.Duration, error)
}
