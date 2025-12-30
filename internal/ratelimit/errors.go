package ratelimit

import (
	"fmt"
)

// ErrRateLimitExceeded is returned when the rate limit is exceeded
var ErrRateLimitExceeded = fmt.Errorf("rate limit exceeded")
