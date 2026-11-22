package utils

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// NewULID generates a new lexicographically sortable 26-character ULID string
func NewULID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

