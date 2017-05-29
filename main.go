package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
)

type option struct {
	url    string
	config string
	out    string
}

var opt option

func init() {
	flag.StringVar(&opt.url, "url", "", "")
	flag.StringVar(&opt.config, "config", "", "")
	flag.StringVar(&opt.out, "out", "", "")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("invalid argument:", flag.Args())
	}
	if opt.url == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		if opt.config == "" {
			opt.config = strings.Join([]string{u.HomeDir, "dotfiles", "etc", "getter", "getter.conf"}, string(os.PathSeparator))
		}
		b, err := ioutil.ReadFile(opt.config)
		if err != nil {
			log.Fatal(err)
		}
		opt.url = strings.TrimSpace(string(b))
	}
}

func read(msg string) string {
	fmt.Print(msg)
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	if sc.Err() != nil {
		panic(sc.Err())
	}
	return sc.Text()
}

func getter(url string) ([]byte, error) {
	resp, err := http.Get(opt.url)
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

func main() {
	if opt.url == "" {
		opt.url = read("url:>")
	}
	if !strings.HasPrefix(opt.url, "https://") && !strings.HasPrefix(opt.url, "http://") {
		opt.url = "https://" + opt.url
	}

	b, err := getter(opt.url)
	if err != nil {
		log.Fatal(err)
	}

	if opt.out != "" {
		err = ioutil.WriteFile(opt.out, b, 0600)
		if err != nil {
			log.Fatal(err)
		}
	}
}
