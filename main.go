package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

const version = "0.5.5"
const logprefix = "getter "

// default discard
var logger = log.New(ioutil.Discard, logprefix, log.LstdFlags)

// option.conf = defconf
var defconf = func() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(u.HomeDir,
		"dotfiles",
		"etc",
		"getter",
		"conf.json")
}()

/// sub commands

type subCommand interface {
	run() error
	init([]string) error
}

var subcmd subCommand
var subCommandsList = []string{
	`get`,
	`getvalues`,
	`version`,

	// TODO: really need sub-list?
	`sub-list`,
}

// getvalues
type subcmdGetValues struct {
	w io.Writer

	// sub flags
	flag  *flag.FlagSet
	fpath string
	elem  string
	attr  string
}

func (sub *subcmdGetValues) init(args []string) error {
	sub.flag = flag.NewFlagSet(args[0], flag.ExitOnError)
	sub.flag.StringVar(&sub.fpath, "file", "", "specify html file path")
	sub.flag.StringVar(&sub.fpath, "f", "", "alias of file")
	sub.flag.StringVar(&sub.elem, "elem", "", "specify search element")
	sub.flag.StringVar(&sub.elem, "e", "", "alias of elem")
	sub.flag.StringVar(&sub.attr, "attr", "", "specify search attribute")
	sub.flag.StringVar(&sub.attr, "a", "", "alias of attr")
	sub.w = os.Stdout
	sub.flag.Parse(args[1:])
	if sub.flag.NArg() != 0 {
		return fmt.Errorf("subcmd: invalid argument:%+v", sub.flag.Args())
	}
	return nil
}
func (sub *subcmdGetValues) run() error {
	b, err := ioutil.ReadFile(sub.fpath)
	if err != nil {
		return err
	}
	values, err := getValues(b, sub.elem, sub.attr)
	if err != nil {
		return err
	}
	for _, s := range values {
		fmt.Fprintln(sub.w, s)
	}
	return nil
}

// simple get
type subcmdGet struct {
	w    io.Writer
	logw io.Writer

	// sub flags
	flag *flag.FlagSet
	url  string
	out  string
	log  bool
}

func (sub *subcmdGet) init(args []string) error {
	sub.w = os.Stdout
	sub.logw = os.Stderr
	sub.flag = flag.NewFlagSet(args[0], flag.ExitOnError)
	sub.flag.StringVar(&sub.url, "url", "", "specify target url")
	sub.flag.StringVar(&sub.out, "out", "", "specify output file")
	sub.flag.BoolVar(&sub.log, "log", false, "output log of get")
	sub.flag.Parse(args[1:])
	if sub.flag.NArg() != 0 {
		return fmt.Errorf("subcmd: invalid argument:%+v", sub.flag.Args())
	}
	return nil
}
func (sub *subcmdGet) run() error {
	if sub.log {
		logger.SetOutput(sub.logw)
	}
	b, err := getter(sub.url)
	if err != nil {
		return err
	}
	if sub.out != "" {
		err = ioutil.WriteFile(sub.out, b, 0666)
		if err != nil {
			return err
		}
		fmt.Fprintln(sub.w, sub.out)
	} else {
		fmt.Fprintln(sub.w, string(b))
	}
	return nil
}

// output of subcmd list
type subcmdList struct {
	w io.Writer

	flag *flag.FlagSet
}

func (sub *subcmdList) init(args []string) error {
	sub.w = os.Stdout
	sub.flag = flag.NewFlagSet(args[0], flag.ExitOnError)
	sub.flag.Parse(args[1:])
	if sub.flag.NArg() != 0 {
		return fmt.Errorf("subcmd: invalid argument:%+v", sub.flag.Args())
	}
	return nil
}
func (sub *subcmdList) run() error {
	for _, s := range subCommandsList {
		fmt.Fprintln(sub.w, s)
	}
	return nil
}

/// getter

func getter(url string) ([]byte, error) {
	// TODO: make flag for specify time of timeout? if need then do that
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
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
	logger.Printf("http header Trailer:%+v\n", resp.Trailer)
	logger.Printf("header:%+v\n", resp.Header)
	logger.Printf("proto:%+v\n", resp.Proto)
	logger.Printf("request:%+v\n", resp.Request)
	logger.Printf("status:%+v\n", resp.Status)
	if resp.TLS != nil {
		logger.Printf("TLS Mutual:%+v\n", resp.TLS.NegotiatedProtocolIsMutual)
		logger.Printf("TLS HandshakeComplete:%+v\n", resp.TLS.HandshakeComplete)
	} else {
		logger.Println("TLS is nil")
	}
	return ioutil.ReadAll(resp.Body)
}

func getValues(b []byte, targetElem, targetAttr string) ([]string, error) {
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var f func(*html.Node)
	var values []string
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == targetElem {
			for _, v := range n.Attr {
				if v.Key == targetAttr {
					values = append(values, v.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return values, nil
}

// json info
type crawlInfo struct {
	URL    string `json:"url"`
	Elem   string `json:"elem"`
	Attr   string `json:"attr"`
	Grep   string `json:"grep"`
	Out    string `json:"out"`
	Outdir string `json:"outdir"`
}

type crawl struct {
	infos []crawlInfo
	delay bool
}

// recuresive get and save
func (c *crawl) do() (string, error) {
	rand.Seed(time.Now().UnixNano())
	var value string
	var lastWrite string
	var url string
	for i, info := range c.infos {
		logger.SetPrefix(fmt.Sprintf(logprefix+"[%v]: ", i))
		url = info.URL
		if value != "" {
			logger.Println("join value:", value)
			url = url + "/" + value
		}
		logger.Println("GET URL:" + url)
		b, err := getter(url)
		if err != nil {
			return "", err
		}
		switch {
		case info.Out != "":
			if err = ioutil.WriteFile(info.Out, b, 0666); err != nil {
				return "", err
			}
			lastWrite = info.Out
		case info.Outdir != "":
			fpath := filepath.Join(info.Outdir, filepath.Base(url))
			if err = ioutil.WriteFile(fpath, b, 0666); err != nil {
				return "", err
			}
			lastWrite = fpath
		}
		var values []string
		if info.Elem != "" && info.Attr != "" {
			values, err = getValues(b, info.Elem, info.Attr)
			if err != nil {
				return "", err
			}
		}
		logger.Println("values:", values)
		switch {
		case info.Grep != "":
			for _, v := range values {
				if strings.Contains(v, info.Grep) {
					value = filepath.Base(v)
					break
				}
			}
		case len(values) != 0:
			value = values[0]
		default:
			value = ""
		}
		if c.delay && i+1 != len(c.infos) {
			delay := time.Duration(10 + rand.Int63n(10))
			logger.Println("delay:", delay)
			time.Sleep(time.Second * delay)
		}
	}
	return lastWrite, nil
}

/// run

func run(w io.Writer) error {
	if opt.version {
		fmt.Fprintf(w, "version %s\n", version)
		return nil
	}
	if opt.list {
		fmt.Fprintf(w, "You can get help: %s [subcmd] --help\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "List Subcommands:\n\t%s\n", strings.Join(subCommandsList, "\n\t"))
		return nil
	}
	if opt.template {
		templInfos := []crawlInfo{
			{
				URL:    "https://",
				Elem:   "specify search of element for next get",
				Attr:   "specify search of attribute for next get",
				Grep:   "specify grep word for pick the value, is join next url",
				Out:    "/path/to/out/file",
				Outdir: "/path/to/out/dir/",
			},
			{
				URL: "specify final get url",
				Out: "specify final output",
			},
		}
		b, err := json.MarshalIndent(templInfos, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(b))
		return nil
	}
	switch {
	case opt.logdrop:
		logger.SetOutput(ioutil.Discard)
	case opt.logfile != "":
		// override
		logfile, err := os.Create(opt.logfile)
		if err != nil {
			return err
		}
		// do not close
		//defer logfile.Close()
		logger.SetOutput(logfile)
	default:
		logger.SetOutput(os.Stderr)
	}
	logger.Println("conf:", opt.conf)
	b, err := ioutil.ReadFile(opt.conf)
	if err != nil {
		return err
	}
	c := &crawl{delay: true}
	if err := json.Unmarshal(b, &c.infos); err != nil {
		return err
	}
	outpath, err := c.do()
	if err != nil {
		return err
	}
	fmt.Fprintln(w, outpath)
	return nil
}

type option struct {
	conf     string
	version  bool
	logdrop  bool
	logfile  string
	template bool
	list     bool

	// add delay time.Duration?
}

var opt option

func init() {
	flag.BoolVar(&opt.version, "version", false, "")
	flag.BoolVar(&opt.logdrop, "logdrop", false, "log dropout")
	flag.StringVar(&opt.conf, "conf", defconf, "specify path to configuration json file")
	flag.StringVar(&opt.logfile, "logfile", "", "specify path to output log file")
	flag.BoolVar(&opt.template, "template", false, "output template configuration json file")
	flag.BoolVar(&opt.list, "list", false, "list subcomamnds")
	flag.Parse()
	if flag.NArg() == 0 {
		return
	}
	switch flag.Arg(0) {
	case "getvalues":
		subcmd = &subcmdGetValues{}
	case "get":
		subcmd = &subcmdGet{}
	case "sub-list":
		subcmd = &subcmdList{}
	case "version":
		if flag.NArg() == 1 {
			opt.version = true
			break
		}
		fallthrough
	default:
		fmt.Fprintf(os.Stderr, "invalid argument:%+v\n", flag.Args())
		os.Exit(1)
	}
	if subcmd != nil {
		if err := subcmd.init(flag.Args()); err != nil {
			fmt.Fprintf(os.Stderr, "initialize error: %+v", err)
			os.Exit(1)
		}
	}
}

func main() {
	switch {
	case subcmd != nil:
		if err := subcmd.run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	default:
		if err := run(os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
