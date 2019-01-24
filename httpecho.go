package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port := 80
	if env := os.Getenv("PORT"); env != "" {
		parsed, err := strconv.Atoi(env)
		if err != nil {
			panic(err)
		}
		port = parsed
	}
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
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
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
