package rest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cognicraft/timeutil"
	"github.com/cognicraft/uuid"
)

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		in:      bufio.NewScanner(r),
		symbols: map[string]string{},
		logf:    logf,
	}
}

// https://www.w3.org/Protocols/rfc2616/rfc2616-sec5.html
type Scanner struct {
	in      *bufio.Scanner
	symbols map[string]string
	err     error
	line    string
	request *http.Request
	logf    func(string, ...interface{})
}

func (s *Scanner) Scan() bool {
	for s.scanLine() {
		if isRequestLine(s.line) {
			// begining
			ok := s.scanRequestLine()
			if !ok {
				return false
			}
			ok = s.scanHeaders()
			if !ok {
				return false
			}
			ok = s.scanBody()
			if !ok {
				return false
			}
			return true
		}
	}
	return false
}

func (s *Scanner) scanLine() bool {
	ok := s.in.Scan()
	if !ok {
		return false
	}
	line := s.in.Text()
	if isComment(line) {
		// comments are be skipped
		return s.scanLine()
	} else if isSymbolDefinition(line) {
		s.scanSymbolDefinition(line)
		return s.scanLine()
	}
	s.line = s.replacePlaceholders(line)
	return true
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "#")
}

func isSymbolDefinition(line string) bool {
	return strings.HasPrefix(line, "@")
}

func isRequestLine(line string) bool {
	if len(line) == 0 {
		return false
	}
	method := strings.ToUpper(strings.Fields(line)[0])
	switch method {
	case http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace:
		return true
	default:
		return false
	}
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) Request() *http.Request {
	return s.request
}

func (s *Scanner) scanSymbolDefinition(line string) bool {
	def := line[1:]
	ps := strings.Split(def, "=")
	if len(ps) != 2 {
		return false
	}
	sym := strings.TrimSpace(ps[0])
	val := strings.TrimSpace(ps[1])
	s.symbols[sym] = val
	return true
}

func (s *Scanner) scanRequestLine() bool {
	fields := strings.Fields(s.line)

	method := "GET"
	rawURL := "/"
	httpVersion := "HTTP/1.1"

	if len(fields) == 1 {
		// only provides request url
		rawURL = fields[0]
	} else {
		method = fields[0]
		rawURL = fields[1]
	}
	var err error
	r, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		s.err = err
		return false
	}
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
	}

	major, minor, ok := http.ParseHTTPVersion(httpVersion)
	if !ok {
		s.err = fmt.Errorf("unknown http version: %s", httpVersion)
		return false
	}
	r.Proto = httpVersion
	r.ProtoMajor = major
	r.ProtoMinor = minor

	s.line = ""
	s.request = r
	return true
}

func (s *Scanner) scanHeaders() bool {
	for s.scanLine() {
		if strings.TrimSpace(s.line) == "" {
			return true
		}
		lcps := strings.Split(s.line, ":")
		key := strings.TrimSpace(lcps[0])
		value := strings.TrimSpace(strings.Join(lcps[1:], ":"))
		s.request.Header.Add(key, value)
		if key == "Host" {
			s.request.URL.Host = value
		}
	}
	return true
}

func (s *Scanner) scanBody() bool {
	// TODO: handle different types of content
	var buf bytes.Buffer
	for s.scanLine() {
		if strings.TrimSpace(s.line) == "" {
			break
		}
		buf.WriteString(s.line)
		buf.WriteString("\n")
	}
	s.request.Body = ioutil.NopCloser(&buf)
	return true
}

func (s *Scanner) replacePlaceholders(line string) string {
	out := ""
	tail := line
	// use functions
	i := strings.Index(tail, "{{")
	for i >= 0 {
		out = out + tail[:i]
		j := strings.Index(tail, "}}")
		if j < 0 {
			s.logf("[ERROR]: no end: %s", tail)
			return out + tail
		}
		verb := tail[i+2 : j]
		out = out + s.replaceVerb(verb)
		tail = tail[j+2:]
		i = strings.Index(tail, "{{")
	}
	out = out + tail
	return out
}

func (s *Scanner) replaceVerb(verb string) string {
	fields := strings.Fields(verb)
	if len(fields) == 0 {
		return ""
	}
	key := fields[0]
	args := fields[1:]
	switch key {
	case "$uuid":
		uuid := uuid.MakeV4()
		if len(args) > 0 {
			switch args[0] {
			case "short":
				return strings.Replace(uuid, "-", "", -1)
			}
		}
		return uuid
	case "$time":
		t := time.Now().UTC()
		switch len(args) {
		case 1:
			zone := args[0]
			loc, err := time.LoadLocation(zone)
			if err != nil {
				loc = time.UTC
			}
			return t.In(loc).Format(time.RFC3339Nano)
		case 2:
			zone := args[0]
			loc, err := time.LoadLocation(zone)
			if err != nil {
				loc = time.UTC
			}
			format := args[1]
			return t.In(loc).Format(timeutil.Layout(format))
		}
		return t.Format(time.RFC3339Nano)
	default:
		if val, ok := s.symbols[key]; ok {
			return val
		}
		return ""
	}
}
