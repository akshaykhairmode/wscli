package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

type Flag struct {
	connectURL  string
	auth        string
	headers     []string
	origin      string
	execute     []string
	wait        time.Duration
	subProtocol []string
	proxy       string

	showPingPong              bool
	isSlash                   bool
	noCertificateCheck        bool
	version                   bool
	verbose                   bool
	noColor                   bool
	shouldShowResponseHeaders bool
	isJSONPrettyPrint         bool
	isBinary                  bool
	isGzipResponse            bool

	help          bool
	isInteractive bool
	isSTDin       bool // read from stdin, cannot send messages to the server other than what is in the stdin

	tls TLS

	mux sync.RWMutex
}

type TLS struct {
	CA         string
	Cert       string
	Key        string
	Passphrase string
}

var Flags *Flag

func init() {
	Flags = Get()
}

func Get() *Flag {
	cfg := Flag{
		mux: sync.RWMutex{},
	}

	pflag.BoolVarP(&cfg.help, "help", "h", false, "print help section")
	pflag.BoolVar(&cfg.isSlash, "slash", false, "pass true if you want to enable slash commands")
	pflag.BoolVarP(&cfg.noCertificateCheck, "no-check", "n", false, "pass true if you want to disable certificate check")
	pflag.BoolVarP(&cfg.showPingPong, "show-ping-pong", "P", false, "pass true if you want to see ping-pong messages")
	pflag.BoolVarP(&cfg.version, "version", "V", false, "print the version")
	pflag.BoolVarP(&cfg.verbose, "verbose", "v", false, "prints the debug logs")
	pflag.BoolVar(&cfg.noColor, "no-color", false, "pass true if you want to disable color output")
	pflag.BoolVarP(&cfg.shouldShowResponseHeaders, "response", "r", false, "pass true if you want to see the http response headers from the server")
	pflag.BoolVarP(&cfg.isSTDin, "stdin", "i", false, "pass true if you want to read from stdin")
	pflag.BoolVar(&cfg.isJSONPrettyPrint, "jspp", false, "pass true if you want to parse the response into json and pretty print")
	pflag.BoolVarP(&cfg.isBinary, "binary", "b", false, "pass if you are sending binary data.")
	pflag.BoolVar(&cfg.isGzipResponse, "gzipr", false, "pass if you are receiving gzip data and need it decoded and printed")

	pflag.StringVarP(&cfg.connectURL, "connect", "c", "", "pass the connection url for the websocket")
	pflag.StringVar(&cfg.proxy, "proxy", "", "pass the proxy url")
	pflag.StringVar(&cfg.auth, "auth", "", "pass the HTTP basic auth")
	pflag.StringSliceVarP(&cfg.headers, "header", "H", []string{}, "pass headers in key:value format, use -H multiple times to pass multiple values, commas also work")
	pflag.StringVarP(&cfg.origin, "origin", "o", "", "optional, pass the origin for the websocket connection")
	pflag.StringSliceVarP(&cfg.execute, "execute", "x", []string{}, "optional, pass the command to execute")
	pflag.DurationVarP(&cfg.wait, "wait", "w", 0, "optional, pass the wait time after executing the command, example 1s, 1m, 1h")
	pflag.StringSliceVarP(&cfg.subProtocol, "sub-protocol", "s", []string{}, "optional, pass the sub-protocol for the websocket connection")

	pflag.StringVar(&cfg.tls.CA, "ca", "", "optional, pass the CA certificate file")
	pflag.StringVar(&cfg.tls.Cert, "cert", "", "optional, pass the client certificate file")
	pflag.StringVar(&cfg.tls.Key, "key", "", "optional, pass the certificate key file")
	pflag.Parse()

	if cfg.help {
		pflag.Usage()
		os.Exit(0)
	}

	if pflag.NFlag() <= 0 {
		cfg.isInteractive = true
	}

	if cfg.noColor {
		color.NoColor = true
	}

	return &cfg
}

func (c *Flag) String() string {
	var sb strings.Builder

	sb.WriteString("Config:\n")
	sb.WriteString(fmt.Sprintf("  connectURL: %s\n", c.connectURL))
	sb.WriteString(fmt.Sprintf("  auth: %s\n", c.auth))
	sb.WriteString(fmt.Sprintf("  headers: %v\n", c.headers))
	sb.WriteString(fmt.Sprintf("  origin: %s\n", c.origin))
	sb.WriteString(fmt.Sprintf("  execute: %v\n", c.execute))
	sb.WriteString(fmt.Sprintf("  wait: %s\n", c.wait))
	sb.WriteString(fmt.Sprintf("  subProtocol: %v\n", c.subProtocol))
	sb.WriteString(fmt.Sprintf("  proxy: %s\n", c.proxy))

	sb.WriteString(fmt.Sprintf("  showPingPong: %t\n", c.showPingPong))
	sb.WriteString(fmt.Sprintf("  isSlash: %t\n", c.isSlash))
	sb.WriteString(fmt.Sprintf("  noCertificateCheck: %t\n", c.noCertificateCheck))
	sb.WriteString(fmt.Sprintf("  version: %t\n", c.version))
	sb.WriteString(fmt.Sprintf("  verbose: %t\n", c.verbose))
	sb.WriteString(fmt.Sprintf("  noColor: %t\n", c.noColor))
	sb.WriteString(fmt.Sprintf("  shouldShowResponseHeaders: %t\n", c.shouldShowResponseHeaders))
	sb.WriteString(fmt.Sprintf("  isJSONPrettyPrint: %t\n", c.isJSONPrettyPrint))
	sb.WriteString(fmt.Sprintf("  isBinary: %t\n", c.isBinary))
	sb.WriteString(fmt.Sprintf("  isGzipResponse: %t\n", c.isGzipResponse))

	sb.WriteString(fmt.Sprintf("  help: %t\n", c.help))
	sb.WriteString(fmt.Sprintf("  isInteractive: %t\n", c.isInteractive))
	sb.WriteString(fmt.Sprintf("  isSTDin: %t\n", c.isSTDin))

	// You may want to customize how TLS is printed
	sb.WriteString(fmt.Sprintf("  tls: %+v\n", c.tls))

	return sb.String()
}
