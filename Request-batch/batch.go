package requestbatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Request holds information to create a new HTTP
type Request struct {
	Method  string
	URL     string
	Headers map[string][]string
	Body    map[string]string
}

// batch holds a slice(array) of request to be sent concurrently and in parallel

type batch struct {
	Request  []*http.Request
	response map[string]struct {
		Index    int
		Responce http.Response
		Err      error
	}
}

// Batch is a constructor function for the batch struct.
func Batch(requests []Request) (*batch, error) {
	batch := batch{}
	err := batch.SetRequests(requests)
	return &batch, err
}

/**
 * Helper function that easily explains errors by logging them in a readable format
 * Takes in the runtime caller information and displays it in the format below:
 * [Jan-02-06 3:04pm] Error Warning: example.go main() line:9 Error invalid argument used by...
 */
func Explain(err error) {
	if err != nil {
		pc, file, line, _ := runtime.Caller(1)
		function := strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), ".")
		fmt.Println("[" + time.Now().Format("Jan-02-06 3:04pm") + "] Error Warning:" + file + " " + function + "() line:" + strconv.Itoa(line) + " " + err.Error())
	}
}

// SetRequests is a setter function for batch's Requests field.
func (b *batch) SetRequests(requests []Request) error {
	reqs := []*http.Request{}
	for _, req := range requests {
		body, err := json.Marshal(req.Body)
		if err != nil {
			return err
		}
		request, err := http.NewRequest(
			req.Method,
			req.URL,
			bytes.NewBuffer(body),
		)
		if err != nil {
			return err
		}
		request.Header = req.Headers
		reqs = append(reqs, request)
	}
	b.Requests = reqs
	return nil
}
