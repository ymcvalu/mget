package mget

import (
	"bufio"
	"context"
	"fmt"
	"github.com/edsrzf/mmap-go"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"runtime"
	"time"
)

type task struct {
	url  string
	path string
	mp   mmap.MMap
	fs   int64
	subs []*subTask

	wg  *errgroup.Group
	ctx context.Context
}

func NewTask(url, path string, nw int) (*task, error) {
	if nw <= 0 {
		nw = runtime.NumCPU()
	}

	sz, multi, err := getFileInfo(url)
	if err != nil {
		return nil, err
	}

	if !multi {
		nw = 1
	}

	t := &task{
		url:  url,
		path: path,
		fs:   sz,
	}
	t.wg, t.ctx = errgroup.WithContext(context.Background())

	if err = func() error {
		for {
			_, err := os.Stat(t.path)
			if os.IsNotExist(err) {
				break
			} else if err != nil {
				return err
			}
			t.path = t.path + "_1"
		}

		fd, err := os.Create(t.path)
		if err != nil {
			return err
		}

		defer fd.Close()

		err = fd.Truncate(t.fs)
		if err != nil {
			return err
		}

		t.mp, err = mmap.Map(fd, mmap.RDWR, 0)
		return err
	}(); err != nil {
		return nil, err
	}

	if nw > 1 {
		var (
			off         int64 = 0
			chuckSz           = t.fs / int64(nw)
			lastChuckSz       = t.fs - int64(nw-1)*chuckSz
		)

		t.subs = make([]*subTask, 0, nw)
		for i := 0; i < nw; i++ {
			sz := chuckSz
			if i == nw-1 {
				sz = lastChuckSz
			}

			t.subs = append(t.subs, &subTask{
				id:   i + 1,
				task: t,
				size: sz,
				off:  off,
				buf:  t.mp[off : off+sz],
			})

			off += sz
		}
	} else {
		t.subs = []*subTask{
			{
				id:   1,
				task: t,
				size: t.fs,
				off:  0,
				buf:  t.mp[:],
			},
		}
	}

	return t, nil
}

func (t *task) exec() error {
	for _, s := range t.subs {
		s := s
		t.wg.Go(func() error {
			return s.exec(t.ctx)
		})
	}

	return t.wg.Wait()
}

const cls = "\r\033[%dA\033[k"

func (t *task) drawDaemon() chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)

		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()

		w := bufio.NewWriter(os.Stdout)
		t.draw(w)
		w.Flush()

		draw := func() {
			fmt.Fprintf(w, cls, len(t.subs)+1)
			t.draw(w)
			w.Flush()
		}

		for {
			select {
			case <-t.ctx.Done():
				draw()
				return
			case <-ticker.C:
				draw()
			}
		}
	}()

	return ch

}

func (t *task) draw(w io.Writer) {
	var all_dl_sz int64
	for _, sub := range t.subs {
		dl_sz := sub.progress()
		all_dl_sz += dl_sz
		sub.drawProcess(w)
	}
	draw(w, "-total-", t.fs, all_dl_sz)
}
