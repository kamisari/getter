package main

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"testing"
)

func TestGetter(t *testing.T) {
	b, err := getter("http://localhost:6060")
	if err != nil {
		t.Fatal(err)
	}
	z := html.NewTokenizer(bytes.NewReader(b))
	var tags []string
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			for _, s := range tags {
				t.Log(s)
			}
			t.Logf("%+v", tt)
			t.Logf("%+v", z.Err())
			break
		}
		if tt == html.StartTagToken {
			t.Logf("%+v,%v", tt, string(z.Raw()))
		} else {
			tags = append(tags, fmt.Sprintf("%s,\tRaw:%s", tt.String(), string(z.Raw())))
		}
	}
}
