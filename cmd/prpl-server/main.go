package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/captaincodeman/prpl-server-go"
	"github.com/go-chi/chi/middleware"
)

var (
	help         bool
	version      bool
	host         string
	port         int
	root         string
	config       string
	httpRedirect bool
)

func init() {
	timestamp := time.Now().Format("20060102150405")

	flag.BoolVar(&help, "help", false, "Print this help text.")
	flag.BoolVar(&version, "version", false, "Print the installed version.")
	flag.StringVar(&host, "host", "127.0.0.1", "Listen on this hostname (default 127.0.0.1).")
	flag.IntVar(&port, "port", 8080, "Listen on this port; 0 for random (default 8080).")
	flag.StringVar(&root, "root", ".", `Serve files relative to this directory (default ".").`)
	flag.StringVar(&config, "config", "", `JSON configuration file (default "<root>/polymer.json" if exists).`)
	flag.BoolVar(&httpRedirect, "http-redirect", false, "Redirect HTTP requests to HTTPS with a 301. Assumes same hostname and default port (443). Trusts X-Forwarded-* headers for detecting protocol and hostname.")
}

func main() {
	flag.Parse()

	if help {
		flag.PrintDefaults()
		return
	}

	if version {
		fmt.Printf("Version 0.0.1/n")
		return
	}

	if host == "" {
		fmt.Printf("invalid --host")
		return
	}

	if port == 0 {
		fmt.Printf("invalid --port")
		return
	}

	if root == "" {
		fmt.Printf("invalid --root")
		return
	}

	m, _ := prpl.New(
		prpl.WithRoot(http.Dir(root)),
		prpl.WithConfigFile(config),
	)

	var h http.Handler

	h = m
	h = middleware.Recoverer(h)
	h = middleware.Logger(h)
	h = middleware.DefaultCompress(h)

	// TODO: graceful shutdown
	// TODO: redirect to https (auto cert?)
	// TODO: handle X-Forwarded- headers

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	http.ListenAndServe(addr, h)
}
