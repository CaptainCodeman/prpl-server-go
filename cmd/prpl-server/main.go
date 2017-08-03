package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/captaincodeman/prpl-server-go"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
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
		prpl.Root(http.Dir(root)),
		prpl.ConfigFile(config),
	)

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.Use(gzip.Gzip(gzip.DefaultCompression))
	n.UseHandler(m)

	// TODO: graceful shutdown
	// TODO: redirect to https (auto cert?)
	// TODO: handle X-Forwarded- headers

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	http.ListenAndServe(addr, n)
}
