package rest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
)

func ParseOne(r io.Reader) (*http.Request, error) {
	p := NewScanner(r)
	p.Scan()
	return p.Request(), p.Err()
}

func ParseAll(r io.Reader) <-chan *http.Request {
	out := make(chan *http.Request, 1)
	go func() {
		defer close(out)
		s := NewScanner(r)
		for s.Scan() {
			out <- s.Request()
		}
	}()
	return out
}

func ParseOneString(raw string) (*http.Request, error) {
	p := NewScanner(strings.NewReader(raw))
	ok := p.Scan()
	if !ok {
		return nil, p.Err()
	}
	return p.Request(), nil
}

func ParseAllString(raw string) []*http.Request {
	requests := []*http.Request{}
	p := NewScanner(strings.NewReader(raw))
	for p.Scan() {
		requests = append(requests, p.Request())
	}
	return requests
}

func DumpRequest(r *http.Request) {
	bs, _ := httputil.DumpRequest(r, true)
	fmt.Printf("%s\n", string(bs))
}

func DumpResponse(r *http.Response) {
	bs, _ := httputil.DumpResponse(r, true)
	fmt.Printf("%s\n", string(bs))
}

func logf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	fmt.Printf(format, args...)
}
