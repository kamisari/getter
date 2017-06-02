package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const version = "0.0.0"

type option struct {
	conf    string
	version bool
	logdrop bool
	// TODO: impl depth
	depth uint

	/// TODO: be graceful

	url1  string
	out1  string
	elem1 string
	attr1 string

	url2  string
	out2  string
	elem2 string
	attr2 string

	url3    string
	grep3   string
	out3    string // have priority
	outdir3 string
}

var opt option

func init() {
	flag.BoolVar(&opt.version, "version", false, "")
	flag.BoolVar(&opt.logdrop, "logdrop", false, "")
	// TODO: impl depth
	flag.UintVar(&opt.depth, "depth", 0, "")
	flag.StringVar(&opt.url1, "url", "", "")
	flag.StringVar(&opt.out1, "out", "", "")
	flag.StringVar(&opt.elem1, "elem", "", "")
	flag.StringVar(&opt.attr1, "attr", "", "")
	flag.StringVar(&opt.conf, "conf", "", "")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("invalid argument:", flag.Args())
	}
	if opt.version {
		fmt.Printf("version %s\n", version)
		os.Exit(0)
	}
	if opt.conf == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		opt.conf = strings.Join([]string{u.HomeDir,
			"dotfiles",
			"etc",
			"getter",
			"getter.conf"}, string(os.PathSeparator))
	}
	b, err := ioutil.ReadFile(opt.conf)
	if err != nil {
		log.Fatal(err)
	}
	confList := strings.Split(strings.TrimSpace(string(b)), "\n")
	for _, s := range confList {
		switch {
		case strings.HasPrefix(s, "depth="):
			// TODO: impl
			if opt.depth == 0 {
				i, err := strconv.Atoi(s)
				if err != nil || i <= 0 {
					continue
				}
				opt.depth = uint(i)
			}
		case strings.HasPrefix(s, "url1="):
			if opt.url1 == "" {
				opt.url1 = strings.TrimSpace(strings.TrimPrefix(s, "url1="))
			}
		case strings.HasPrefix(s, "out1="):
			if opt.out1 == "" {
				opt.out1 = strings.TrimSpace(strings.TrimPrefix(s, "out1="))
			}
		case strings.HasPrefix(s, "elem1="):
			if opt.elem1 == "" {
				opt.elem1 = strings.TrimSpace(strings.TrimPrefix(s, "elem1="))
			}
		case strings.HasPrefix(s, "attr1="):
			if opt.attr1 == "" {
				opt.attr1 = strings.TrimSpace(strings.TrimPrefix(s, "attr1="))
			}

		case strings.HasPrefix(s, "url2="):
			if opt.url2 == "" {
				opt.url2 = strings.TrimSpace(strings.TrimPrefix(s, "url2="))
			}
		case strings.HasPrefix(s, "out2="):
			if opt.out2 == "" {
				opt.out2 = strings.TrimSpace(strings.TrimPrefix(s, "out2="))
			}
		case strings.HasPrefix(s, "elem2="):
			if opt.elem2 == "" {
				opt.elem2 = strings.TrimSpace(strings.TrimPrefix(s, "elem2="))
			}
		case strings.HasPrefix(s, "attr2="):
			if opt.attr2 == "" {
				opt.attr2 = strings.TrimSpace(strings.TrimPrefix(s, "attr2="))
			}

		case strings.HasPrefix(s, "url3="):
			if opt.url3 == "" {
				opt.url3 = strings.TrimSpace(strings.TrimPrefix(s, "url3="))
			}
		case strings.HasPrefix(s, "grep3="):
			if opt.grep3 == "" {
				opt.grep3 = strings.TrimSpace(strings.TrimPrefix(s, "grep3="))
			}
		case strings.HasPrefix(s, "out3="):
			if opt.out3 == "" {
				opt.out3 = strings.TrimSpace(strings.TrimPrefix(s, "out3="))
			}
		case strings.HasPrefix(s, "outdir3="):
			if opt.outdir3 == "" {
				opt.outdir3 = strings.TrimSpace(strings.TrimPrefix(s, "outdir3="))
			}
		}
	}
}

// TODO: be graceful
func getter(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println("http header", resp.Trailer)
	log.Println("header:", resp.Header)
	log.Println("proto:", resp.Proto)
	log.Println("request:", resp.Request)
	log.Println("status:", resp.Status)
	if resp.TLS != nil {
		log.Println("TLS Mutual:", resp.TLS.NegotiatedProtocolIsMutual)
		log.Println("TLS HandshakeComplete:", resp.TLS.HandshakeComplete)
	} else {
		log.Println("TLS is nil")
	}
	return ioutil.ReadAll(resp.Body)
}
func getValues(b []byte, targetElem, targetAttr string) ([]string, error) {
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var f func(*html.Node)
	var valuse []string
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == targetElem {
			for _, v := range n.Attr {
				if v.Key == targetAttr {
					valuse = append(valuse, v.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return valuse, nil
}

func main() {
	var b []byte
	var values []string
	var err error
	rand.Seed(time.Now().UnixNano())
	if opt.logdrop {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	log.Println("default conf:", opt.conf)

	/// TODO: url[1..3] be graceful

	// url1
	if opt.url1 != "" {
		log.SetPrefix("[1] ")
		log.Println("GET:", opt.url1)
		b, err = getter(opt.url1)
		if err != nil {
			log.Fatal(err)
		}
		if opt.elem1 != "" && opt.attr1 != "" {
			values, err = getValues(b, opt.elem1, opt.attr1)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("values:", values)
		}
		if opt.out1 != "" {
			err = ioutil.WriteFile(opt.out1, b, 0600)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("outfile:", opt.out1)
		}
		delay := time.Duration(2 + rand.Int63n(3))
		log.Println("delay:", delay)
		time.Sleep(time.Second * delay)
	}

	// url2
	if opt.url2 != "" && len(values) != 0 {
		url := opt.url2 + "/" + values[0]
		log.SetPrefix("[2] ")
		log.Println("GET:", url)
		b, err = getter(url)
		if err != nil {
			log.Fatal(err)
		}
		if opt.elem2 != "" && opt.attr2 != "" {
			values, err = getValues(b, opt.elem2, opt.attr2)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("values:", values)
		}
		if opt.out2 != "" {
			err = ioutil.WriteFile(opt.out2, b, 0600)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("outfile:", opt.out2)
		}
		delay := time.Duration(2 + rand.Int63n(3))
		log.Println("delay:", delay)
		time.Sleep(time.Second * delay)
	}

	// url3
	if opt.url3 != "" && len(values) != 0 {
		log.SetPrefix("[3] ")
		log.Println("url:", opt.url3)
		out3 := ""
		for _, v := range values {
			if strings.Contains(v, opt.grep3) {
				url := opt.url3 + "/" + filepath.Base(v)
				log.Println("found v:", v)
				log.Println("GET:", url)
				b, err = getter(url)
				if err != nil {
					log.Fatal(err)
				}
				switch {
				case opt.outdir3 != "":
					out3 = filepath.Join(opt.outdir3, filepath.Base(v))
				case opt.out3 != "":
					out3 = opt.out3
				default:
					log.Println("output not specified")
				}
				break
			}
		}
		if out3 != "" {
			log.Println("outfile:", out3)
			err = ioutil.WriteFile(out3, b, 0600)
			if err != nil {
				log.Fatal(err)
			}
			out3, err = filepath.Abs(out3)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(out3)
		}
	}
}
