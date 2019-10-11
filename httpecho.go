package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func getSingleHeader(req *http.Request, name string) (string, bool) {
	hdr := http.CanonicalHeaderKey(name)
	if vals, ok := req.Header[hdr]; ok && len(vals) > 0 {
		return vals[0], len(vals) == 1
	}
	return "", false
}

func echoJSON(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	ret := map[string]interface{}{}
	ret["method"] = req.Method
	ret["path"] = req.URL.Path
	if req.URL.ForceQuery || req.URL.RawQuery != "" {
		ret["query"] = req.URL.RawQuery
	}
	if req.URL.User != nil {
		ret["user"] = req.URL.User
	}
	if len(req.Header) > 0 {
		hdrs := map[string]interface{}{}
		for name, vals := range req.Header {
			if len(vals) == 1 {
				hdrs[name] = vals[0]
			} else {
				hdrs[name] = vals
			}
		}
		ret["headers"] = hdrs
	}
	var body []byte
	info := map[string]interface{}{}
	n, err := req.Body.Read(body)
	tryParsing := false
	switch {
	case err == io.EOF:
		if cl, ok := getSingleHeader(req, "Content-Length"); ok && cl == "0" {
			tryParsing = true
		}
	case err == nil:
		tryParsing = true
	default:
		info["error"] = err.Error()
	}
	if tryParsing {
		rest, readErr := ioutil.ReadAll(req.Body)
		if readErr == nil {
			body = append(body, rest...)
			info["length"] = len(body)
			info["contents"] = string(body)
			parsed := []interface{}{}
			errs := []error{}
			dec := json.NewDecoder(bytes.NewReader(body))
			for {
				more := false
				var i interface{}
				err := dec.Decode(&i)
				switch err {
				case io.EOF:
				case nil:
					parsed = append(parsed, i)
					more = true
				default:
					errs = append(errs, err)
				}
				if !more {
					break
				}
			}
			switch {
			case len(errs) > 0:
				msgs := []string{}
				for _, err := range errs {
					msgs = append(msgs, fmt.Sprintf("%v", err))
				}
				info["errors"] = msgs
			case len(parsed) == 0:
				info["parsed"] = json.RawMessage(`null`)
			case len(parsed) == 1:
				info["parsed"] = parsed[0]
			default:
				info["parsed"] = parsed
			}
			ret["body"] = info
		} else {
			info["read"] = n
			info["error"] = readErr.Error()
		}
	}
	if len(info) > 0 {
		ret["body"] = info
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(ret)
}

func echoPlain(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain; charset=utf8")
	fmt.Fprintf(w, "Method: %s\n", req.Method)
	fmt.Fprintf(w, "Path: %s\n", req.URL.Path)
	if req.URL.ForceQuery || req.URL.RawQuery != "" {
		fmt.Fprintf(w, "Query: %s\n", req.URL.RawQuery)
	}
	if req.URL.User != nil {
		fmt.Fprintf(w, "User: %v\n", req.URL.User)
	}
	if len(req.Header) > 0 {
		io.WriteString(w, "Request Headers:\n")
		for name, vals := range req.Header {
			for _, val := range vals {
				fmt.Fprintf(w, "  %s: %s\n", name, val)
			}
		}
	}

	var body []byte
	if n, err := req.Body.Read(body); err == nil {
		rest, readErr := ioutil.ReadAll(req.Body)
		if readErr == nil {
			body = append(body, rest...)
			fmt.Fprintf(w, "Body: length=%d\n", len(body))
			fmt.Fprintf(w, "Contents:\n")
			if len(body) == 0 {
				fmt.Fprintf(w, "  (empty)\n")
			} else {
				offset := 0
				stride := 0x20
				nIndent := len(fmt.Sprintf("%x", 1+len(body)-stride))
				output := func(section []byte) {
					fmt.Fprintf(w, "  0x%0*x%v\n", nIndent, offset, section)
				}
				for offset < len(body) {
					end := offset + stride
					if end >= len(body) {
						end = len(body)
					}
					output(body[offset:end])
					offset = offset + stride
				}
			}
		} else {
			plural := "s"
			if n == 1 {
				plural = ""
			}
			fmt.Fprintf(w, "Read %d byte%s successfully.\n", n, plural)
			fmt.Fprintf(w, "Error reading body: %s", readErr)
		}
	}
}

func wantsJSON(req *http.Request) bool {
	for _, h := range []string{"Accept", "Content-Type"} {
		vals, found := req.Header[http.CanonicalHeaderKey(h)]
		if !found {
			continue
		}
		for _, v := range vals {
			if strings.Contains(v, "json") {
				return true
			}
		}
	}
	return false
}

type hteeteep struct {
	w    http.ResponseWriter
	also io.Writer
}

func (t hteeteep) Header() http.Header {
	return t.w.Header()
}

func (t hteeteep) Write(b []byte) (int, error) {
	r, e := t.w.Write(b)
	t.also.Write(b)
	return r, e
}

func (t hteeteep) WriteHeader(s int) {
	t.w.WriteHeader(s)
	fmt.Fprintf(t.also, "STATUS: %d\n", s)
}

func main() {
	stdout := true
	quiet := false
	port := 80
	if env := os.Getenv("PORT"); env != "" {
		parsed, err := strconv.Atoi(env)
		if err != nil {
			panic(err)
		}
		port = parsed
	}

	flag.BoolVar(&quiet, "quiet", quiet, "Don't print to stdout, too")

	flag.Parse()

	if quiet {
		stdout = false
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var tee http.ResponseWriter
		if stdout {
			tee = hteeteep{w, os.Stdout}
		}
		switch {
		case wantsJSON(req):
			echoJSON(tee, req)
		default:
			echoPlain(tee, req)
		}
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
