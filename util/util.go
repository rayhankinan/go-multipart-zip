package util

import (
	"fmt"
	"io"
)

func StreamSizeBySeeking(s io.Seeker) (int64, error) {
	currentPosition, err := s.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, fmt.Errorf("getting current offset: %w", err)
	}
	maxPosition, err := s.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("fast-forwarding to end: %w", err)
	}
	_, err = s.Seek(currentPosition, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("returning to prior offset %d: %w", currentPosition, err)
	}
	return maxPosition, nil
}
