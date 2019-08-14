package mget

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
)

type subTask struct {
	id    int
	task  *task
	size  int64
	dl_sz int64
	off   int64
	buf   []byte
}

func (t *subTask) exec(ctx context.Context) error {
	resp, err := getRange(ctx, t.task.url, t.off, t.off+t.size-1)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	buf := t.buf
loop:
	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				t.advance(n)
				break
			}
			return err
		}

		t.advance(n)
		buf = buf[n:]

		select {
		case <-ctx.Done():
			break loop
		default:
		}
	}

	return nil
}

func (t *subTask) progress() int64 {
	return atomic.LoadInt64(&t.dl_sz)
}

func (t *subTask) advance(sz int) {
	atomic.AddInt64(&t.dl_sz, int64(sz))
}

func (t *subTask) percent() float64 {
	return float64(t.progress()) / float64(t.size)
}

func (t *subTask) drawProcess(w io.Writer) {
	draw(w, fmt.Sprintf("task_%02d", t.id), t.size, t.progress())
}
