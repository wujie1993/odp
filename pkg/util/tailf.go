package util

import (
	"context"

	"github.com/hpcloud/tail"
)

func Tailf(ctx context.Context, path string) (<-chan string, error) {
	watcher := make(chan string, 1000)

	tailf, err := tail.TailFile(path, tail.Config{Follow: true, MustExist: true})
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(watcher)
		for {
			select {
			case line, ok := <-tailf.Lines:
				if !ok {
					return
				}
				watcher <- line.Text
			case <-ctx.Done():
				return
			}
		}
	}()
	return watcher, nil
}
