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

	pflag.BoolVarP(&cfg.help, "help", "h", false, "	Display help information.")
	pflag.BoolVar(&cfg.isSlash, "slash", false, "Enable slash commands (Experimental).")
	pflag.BoolVarP(&cfg.noCertificateCheck, "no-check", "n", false, "Disable TLS certificate verification.")
	pflag.BoolVarP(&cfg.showPingPong, "show-ping-pong", "P", false, "Show ping/pong messages.")
	pflag.BoolVarP(&cfg.version, "version", "V", false, "Display version information.")
	pflag.BoolVarP(&cfg.verbose, "verbose", "v", false, "Enable debug logging.")
	pflag.BoolVar(&cfg.noColor, "no-color", false, "Disable colored output.")
	pflag.BoolVarP(&cfg.shouldShowResponseHeaders, "response", "r", false, "Display HTTP response headers from the server.")
	pflag.BoolVar(&cfg.isJSONPrettyPrint, "jspp", false, "Enable JSON pretty printing for responses.")
	pflag.BoolVarP(&cfg.isBinary, "binary", "b", false, "Send hex encoded data to server")
	pflag.BoolVar(&cfg.isGzipResponse, "gzipr", false, "Enable gzip decoding if server messages are gzip-encoded. (Note: Server must send messages as binary.)")

	pflag.StringVarP(&cfg.connectURL, "connect", "c", "", "WebSocket connection URL.")
	pflag.StringVar(&cfg.proxy, "proxy", "", "Use a proxy URL.")
	pflag.StringVar(&cfg.auth, "auth", "", "HTTP Basic Authentication credentials (e.g., username:password).")
	pflag.StringSliceVarP(&cfg.headers, "header", "H", []string{}, "Custom headers (key:value, can be used multiple times).")
	pflag.StringVarP(&cfg.origin, "origin", "o", "", "Specify origin for the WebSocket connection (optional).")
	pflag.StringSliceVarP(&cfg.execute, "execute", "x", []string{}, "Execute a command after connecting (use multiple times for multiple commands).")
	pflag.DurationVarP(&cfg.wait, "wait", "w", 0, "Wait time after command execution (1s, 1m, 1h).")
	pflag.StringSliceVarP(&cfg.subProtocol, "sub-protocol", "s", []string{}, "Specify a sub-protocol for the WebSocket connection (optional, can be used multiple times).")

	pflag.StringVar(&cfg.tls.CA, "ca", "", "Path to the CA certificate file (optional).")
	pflag.StringVar(&cfg.tls.Cert, "cert", "", "Path to the client certificate file (optional).")
	pflag.StringVar(&cfg.tls.Key, "key", "", "Path to the certificate key file (optional).")
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

	cfg.isSTDin = isInputFromPipe()

	return &cfg
}

func isInputFromPipe() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) == 0
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
