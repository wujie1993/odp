package util_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/wujie1993/waves/pkg/util"
)

func TestTailf(t *testing.T) {
	path := "./tailf.test"

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	tailf, err := util.Tailf(ctx, path)
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		for i := 0; i <= 5; i++ {
			if _, err := f.WriteString(time.Now().String() + "\n"); err != nil {
				t.Error(err)
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
		f.Close()
		os.Remove(path)
		cancel()
	}()

	for {
		select {
		case line, ok := <-tailf:
			if !ok {
				return
			}
			t.Log(line)
		case <-ctx.Done():
			return
		}
	}
}
