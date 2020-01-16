package mget

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

func getRange(ctx context.Context, url string, from, to int64) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if from >= 0 && to >= 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", from, to))
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, errors.New(fmt.Sprintf("code: %d, status: %s", resp.StatusCode, resp.Status))
	}

	return resp, nil
}

var contentRangeReg = regexp.MustCompile("bytes .*/[0-9]*")

func getFileInfo(url string) (int64, bool, error) {
	resp, err := getRange(context.Background(), url, 0, 0)
	if err != nil {
		return 0, false, err
	}

	if resp.StatusCode == http.StatusOK {
		_size := resp.Header.Get("Content-Length")
		size, _ := strconv.ParseInt(_size, 10, 64)
		return size, false, nil
	}

	_range := resp.Header.Get("Content-Range")
	if !contentRangeReg.MatchString(_range) {
		return 0, false, nil
	}

	var _size string
	for i := len(_range) - 1; i >= 0; i-- {
		if _range[i] == '/' {
			_size = _range[i+1:]
			break
		}
	}

	size, _ := strconv.ParseInt(_size, 10, 64)
	return size, true, nil
}

func draw(w io.Writer, t string, fs, dl int64) {
	fmt.Fprintf(w, "\033[33m%-8s\033[0m", t)

	percent := dl * 100 / fs
	d := int(50 * percent / 100)
	dd := 50 - d
	fmt.Fprint(w, "\033[36m[")
	for i := 0; i < d; i++ {
		w.Write([]byte{'+'})
	}

	for i := 0; i < dd; i++ {
		w.Write([]byte{'-'})
	}

	fmt.Fprintf(w, "]\033[0m")
	w.Write([]byte(fmt.Sprintf(" \033[34m%3d%% [%s/%s]\033[0m\n", percent, sizeUnit(dl), sizeUnit(fs))))
	return
}

const (
	KB = 1 << 10
	MB = 1 << 20
	GB = 1 << 30
)

func sizeUnit(bs int64) string {
	var ss string
	switch {
	case bs < KB*2/3:
		ss = fmt.Sprintf("%6dB", bs)
	case bs < MB*2/3:
		ss = fmt.Sprintf("%6.2fKB", float64((bs*100)/KB)/100.0)
	case bs < GB*2/3:
		ss = fmt.Sprintf("%6.2fMB", float64((bs*100)/MB)/100.0)
	default:
		ss = fmt.Sprintf("%6.2fGB", float64((bs*100)/GB)/100.0)
	}

	return ss
}
