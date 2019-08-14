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
	w.Write([]byte(fmt.Sprintf(" \033[34m%d%% [%d/%d]\033[0m\n", percent, dl, fs)))
	return
}
