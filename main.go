package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type option struct {
	conf string

	/// TODO: be graceful

	url1   string
	out1   string
	telem1 string
	tattr1 string

	url2   string
	out2   string
	telem2 string
	tattr2 string

	url3    string
	grep3   string
	out3    string // have priority
	outdir3 string
}

var opt option

func init() {
	flag.StringVar(&opt.url1, "url", "", "")
	flag.StringVar(&opt.out1, "out", "", "")
	flag.StringVar(&opt.telem1, "telem", "", "")
	flag.StringVar(&opt.tattr1, "tattr", "", "")
	flag.StringVar(&opt.conf, "conf", "", "")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("invalid argument:", flag.Args())
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
		log.Println("default conf:", opt.conf)
	}
	b, err := ioutil.ReadFile(opt.conf)
	if err != nil {
		log.Fatal(err)
	}
	confList := strings.SplitAfter(strings.TrimSpace(string(b)), "\n")
	for _, s := range confList {
		switch {
		case strings.HasPrefix(s, "url1="):
			if opt.url1 == "" {
				opt.url1 = strings.TrimSpace(strings.TrimPrefix(s, "url1="))
			}
		case strings.HasPrefix(s, "out1="):
			if opt.out1 == "" {
				opt.out1 = strings.TrimSpace(strings.TrimPrefix(s, "out1="))
			}
		case strings.HasPrefix(s, "telem1="):
			if opt.telem1 == "" {
				opt.telem1 = strings.TrimSpace(strings.TrimPrefix(s, "telem1="))
			}
		case strings.HasPrefix(s, "tattr1="):
			if opt.tattr1 == "" {
				opt.tattr1 = strings.TrimSpace(strings.TrimPrefix(s, "tattr1="))
			}

		case strings.HasPrefix(s, "url2="):
			if opt.url2 == "" {
				opt.url2 = strings.TrimSpace(strings.TrimPrefix(s, "url2="))
			}
		case strings.HasPrefix(s, "out2="):
			if opt.out2 == "" {
				opt.out2 = strings.TrimSpace(strings.TrimPrefix(s, "out2="))
			}
		case strings.HasPrefix(s, "telem2="):
			if opt.telem2 == "" {
				opt.telem2 = strings.TrimSpace(strings.TrimPrefix(s, "telem2="))
			}
		case strings.HasPrefix(s, "tattr2="):
			if opt.tattr2 == "" {
				opt.tattr2 = strings.TrimSpace(strings.TrimPrefix(s, "tattr2="))
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

func getter(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
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
	log.SetOutput(os.Stderr)

	// url1
	if opt.url1 != "" {
		log.SetPrefix("[1] ")
		log.Println("GET:", opt.url1)
		if b, err = getter(opt.url1); err != nil {
			log.Fatal(err)
		}
		values, err = getValues(b, opt.telem1, opt.tattr1)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("values:", values)
		log.Println("outfile:", opt.out1)
		if opt.out1 != "" {
			err = ioutil.WriteFile(opt.out1, b, 0600)
			if err != nil {
				log.Fatal(err)
			}
		}
		duray := time.Duration(2 + rand.Int63n(3))
		log.Println("duray:", duray)
		time.Sleep(time.Second * duray)
	}

	// url2
	if opt.url2 != "" && len(values) != 0 {
		log.SetPrefix("[2] ")
		url := opt.url2 + "/" + values[0]
		log.Println("GET:", url)
		b, err = getter(url)
		if err != nil {
			log.Fatal(err)
		}
		values, err = getValues(b, opt.telem2, opt.tattr2)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("values:", values)
		log.Println("outfile:", opt.out2)
		err = ioutil.WriteFile(opt.out2, b, 0600)
		if err != nil {
			log.Fatal(err)
		}
		duray := time.Duration(2 + rand.Int63n(3))
		log.Println("duray:", duray)
		time.Sleep(time.Second * duray)
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
			// TODO: remove it?
			out3, err = filepath.Abs(out3)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(out3)
		}
	}
}
