package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"sync"
)

// Mutex protected increament counter to process the given max urls
type UrlProcessCount struct {
	count     int
	max_count int
	mu        sync.Mutex
}

func NewUrlProcessCount(max_count int) *UrlProcessCount {
	return &UrlProcessCount{max_count: max_count}
}

func (counter *UrlProcessCount) Increase() bool {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	if counter.count < counter.max_count {
		counter.count++
		return true
	}
	return false
}

// Hashing function
func Md5(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return hex.EncodeToString(h.Sum(nil))
}
