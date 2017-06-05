package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var mockT *testing.T
var mockHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello from mock handler")
	const from = "from mock handler:"
	mockT.Log(from, "header:", r.Header)
	mockT.Log(from, "proto:", r.Proto)
	mockT.Log(from, "host:", r.Host)
	mockT.Log(from, "request:", r.RequestURI)
	mockT.Log(from, "tls:", r.TLS)
})

func TestMain(m *testing.M) {
	opt.conf = ""
	os.Exit(m.Run())
}

func TestGetter(t *testing.T) {
	mockT = t
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

func TestGetter_HTTPS(t *testing.T) {
	mockT = t
	ts := httptest.NewTLSServer(mockHandler)
	defer ts.Close()
	b, err := getter(ts.URL)
	if err == nil {
		t.Fatal("expected certificate error but nil")
	}
	t.Logf("error: %+v\n", err)
	t.Logf("return string([]byte): %+v\n", string(b))
}

func TestRun(t *testing.T) {
	var s string
	buf := bytes.NewBufferString(s)
	err := run(buf)
	if err != nil {
		t.Log(buf)
		t.Fatal(err)
	}
	t.Log(buf)
}

func TestCrawl_Do(t *testing.T) {
}

func TestJSON_Marshal(t *testing.T) {
	var c = new(crawl)
	c.infos = append(c.infos, crawlInfo{
		URL:  "http://hello",
		Elem: "next",
		Attr: "dor",
	})
	c.infos = append(c.infos, crawlInfo{
		URL:  "http://world",
		Elem: "end",
		Attr: "of",
		Grep: "di",
	})
	b, err := json.MarshalIndent(c.infos, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", b)
}

func TestJSON_Unmarshal(t *testing.T) {
	b := []byte(`[
  {
    "url": "http://hello",
    "elem": "next",
    "attr": "dor",
    "grep": "",
    "out": "",
    "outdir": ""
  },
  {
    "url": "http://world",
    "elem": "end",
    "attr": "of",
    "grep": "di",
    "out": "",
    "outdir": ""
  }
]`)
	c := new(crawl)
	err := json.Unmarshal(b, &c.infos)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", c.infos)
}
