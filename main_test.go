package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetter(t *testing.T) {
	var mockHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello from mock handler")
	})
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
