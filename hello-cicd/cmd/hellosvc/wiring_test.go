package main

import (
	"hello-cicd/pkg/helloendpoint"
	"hello-cicd/pkg/helloservice"
	"hello-cicd/pkg/hellotransport"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/log"
)

func TestHTTP(t *testing.T) {
	svc := helloservice.New(log.NewNopLogger())
	eps := helloendpoint.New(svc, log.NewNopLogger(), discard.NewHistogram())
	mux := hellotransport.NewHTTPHandler(eps, log.NewNopLogger())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	for _, testcase := range []struct {
		method, url, body, want string
	}{
		{"POST", srv.URL + "/sayHello", `{"a":"Spin"}`, `{"v":"Hello Spin"}`},
		{"GET", srv.URL + "/sayHello", ``, `{"v":"Hello World, this is a basic CICD demo."}`},
	} {
		req, _ := http.NewRequest(testcase.method, testcase.url, strings.NewReader(testcase.body))
		resp, _ := http.DefaultClient.Do(req)
		body, _ := io.ReadAll(resp.Body)
		if want, have := testcase.want, strings.TrimSpace(string(body)); want != have {
			t.Errorf("%s %s %s: want %q, have %q", testcase.method, testcase.url, testcase.body, want, have)
		}
		time.Sleep(2 * time.Second)
	}
}
