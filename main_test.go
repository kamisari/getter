package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const page = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title></title>
</head>
<body>
  hello mock server
</body>
</html>
`

var mockHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "%s", page)
})

func TestMain(m *testing.M) {
	opt.conf = ""
	os.Exit(m.Run())
}

func TestGetter_HTTP(t *testing.T) {
	ts := httptest.NewServer(mockHandler)
	defer ts.Close()
	b, err := getter(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != page {
		t.Errorf("out: %+v\n", string(b))
		t.Errorf("exp: %+v\n", page)
	}
	t.Logf("%+v\n", string(b))
}

// TODO: certificate error
func TestGetter_HTTPS(t *testing.T) {
	ts := httptest.NewTLSServer(mockHandler)
	defer ts.Close()
	b, err := getter(ts.URL)
	// fatal check
	if err == nil {
		t.Fatal("expected certificate error but nil")
	}
	t.Logf("error: %+v\n", err)
	t.Logf("return string([]byte): %+v\n", string(b))
}

// TODO: fatal check
func TestGetValues(t *testing.T) {
	values, err := getValues([]byte(page), "meta", "charset")
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 {
		t.Fatalf("unexpected values: %+v", values)
	}
	if values[0] != "utf-8" {
		t.Fatalf("expected: utf-8 but %+v", values[0])
	}
}

// TODO: certificate error
func TestRun(t *testing.T) {
	var s string
	buf := bytes.NewBufferString(s)
	// fatal check
	if err := run(buf); err == nil {
		t.Error(buf)
		t.Fatal("expected err but nil")
	}
	t.Log("output:", buf)

	ts := httptest.NewTLSServer(mockHandler)
	defer ts.Close()
	jsondata := []byte(`[
  {
    "url": "` + ts.URL + `",
    "elem": "meta",
    "attr": "char-set",
    "grep": "",
    "out": "",
    "outdir": ""
  }
]`)
	f, err := ioutil.TempFile("", "getter_test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(jsondata); err != nil {
		t.Fatal(err)
	}
	opt.conf = f.Name()
	// fatal check certificate error
	if err := run(buf); err == nil {
		t.Error("out put:", buf)
		t.Fatal("expected error but nil")
	}
	t.Log("output:", buf)

	// TODO: sanity check
	//     : implement tls server or skip certificate check
}

// TODO: implementation
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
