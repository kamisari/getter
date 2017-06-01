package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var testURL string

var mockHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello from mock handler")
	fmt.Println("on mock handler")
	fmt.Println("header:", r.Header)
	fmt.Println("proto:", r.Proto)
	fmt.Println("host:", r.Host)
	fmt.Println("request:", r.RequestURI)
	fmt.Println("tls:", r.TLS)
})

func TestMain(m *testing.M) {
	b, err := ioutil.ReadFile("./t/testURL.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	testURL = strings.TrimSpace(string(b))
	os.Exit(m.Run())
}

func TestGetter(t *testing.T) {
	ts := httptest.NewServer(mockHandler)
	defer ts.Close()

	b, err := getter(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != string("hello from mock handler\n") {
		t.Errorf("out: %+v\n", string(b))
	}
	t.Logf("%+v\n", string(b))
}

func TestGetterHTTPS(t *testing.T) {
	ts := httptest.NewTLSServer(mockHandler)
	defer ts.Close()

	b, err := getter(ts.URL)
	if err == nil {
		t.Fatal("expected error but nil")
	}
	t.Logf("%+v\n", string(b))
}

func TestGetterHTTPSServer(t *testing.T) {
	// TODO:
	_, err := getter(testURL)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDefaultHTTPClient(t *testing.T) {
	// TODO:
	resp, err := http.Get(testURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log("header:", resp.Header)
	t.Log("proto:", resp.Proto)
	t.Log("request:", resp.Request)
	t.Log("status:", resp.Status)
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}
