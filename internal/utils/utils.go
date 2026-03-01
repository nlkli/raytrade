package utils

import (
	"context"
	"os"
	"time"
)

func WatchFile(
	ctx context.Context,
	path string,
	duration time.Duration,
) (<-chan []byte, error) {

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	lastModTime := fileInfo.ModTime()

	ticker := time.NewTicker(duration)

	ch := make(chan []byte, 1)

	go func() {
		defer ticker.Stop()
		defer close(ch)

		for {
			select {
			case <-ticker.C:
				fileInfo, err := os.Stat(path)
				if err != nil {
					continue
				}

				modTime := fileInfo.ModTime()

				if modTime.After(lastModTime) {
					b, err := os.ReadFile(path)
					if err != nil {
						continue
					}
					ch <- b
					lastModTime = modTime
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}
