package searcher

import (
	"errors"
	"sync/atomic"
)

var currentIndex = new(int64)
var errHasherDone = errors.New("hasher have done its jobs")

func GetJobNumber() (int, error) {
	n := atomic.AddInt64(currentIndex, Step)
	if n > MaxNumber {
		return 0, errHasherDone
	}

	return int(n), nil
}
