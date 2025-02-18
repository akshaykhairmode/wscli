package config

import (
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

type Config struct {
	ConnectURL  string
	Auth        string
	Headers     []string
	Origin      string
	Execute     []string
	Wait        time.Duration
	SubProtocol []string
	Proxy       string

	ShowPingPong       bool
	IsSlash            bool
	NoCertificateCheck bool
	Version            bool
	Verbose            bool
	NoColor            bool
	Response           bool

	Help          bool
	IsInteractive bool
	Stdin         bool // read from stdin, cannot send messages to the server other than what is in the stdin

	TLS TLS
}

type TLS struct {
	CA         string
	Cert       string
	Key        string
	Passphrase string
}

func Get() Config {
	cfg := Config{}

	pflag.BoolVarP(&cfg.Help, "help", "h", false, "print help section")
	pflag.BoolVar(&cfg.IsSlash, "slash", false, "pass true if you want to enable slash commands")
	pflag.BoolVarP(&cfg.NoCertificateCheck, "no-check", "n", false, "pass true if you want to disable certificate check")
	pflag.BoolVarP(&cfg.ShowPingPong, "show-ping-pong", "P", false, "pass true if you want to see ping-pong messages")
	pflag.BoolVarP(&cfg.Version, "version", "V", false, "print the version")
	pflag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "prints the debug logs")
	pflag.BoolVar(&cfg.NoColor, "no-color", false, "pass true if you want to disable color output")
	pflag.BoolVarP(&cfg.Response, "response", "r", false, "pass true if you want to see the http response headers from the server")
	pflag.BoolVarP(&cfg.Stdin, "stdin", "i", false, "pass true if you want to read from stdin")

	pflag.StringVarP(&cfg.ConnectURL, "connect", "c", "", "pass the connection url for the websocket")
	pflag.StringVar(&cfg.Proxy, "proxy", "", "pass the proxy url")
	pflag.StringVar(&cfg.Auth, "auth", "", "pass the HTTP basic auth")
	pflag.StringSliceVarP(&cfg.Headers, "header", "H", []string{}, "pass headers in key:value format, use -H multiple times to pass multiple values, commas also work")
	pflag.StringVarP(&cfg.Origin, "origin", "o", "", "optional, pass the origin for the websocket connection")
	pflag.StringSliceVarP(&cfg.Execute, "execute", "x", []string{}, "optional, pass the command to execute")
	pflag.DurationVarP(&cfg.Wait, "wait", "w", 0, "optional, pass the wait time after executing the command, example 1s, 1m, 1h")
	pflag.StringSliceVarP(&cfg.SubProtocol, "sub-protocol", "s", []string{}, "optional, pass the sub-protocol for the websocket connection")

	pflag.StringVar(&cfg.TLS.CA, "ca", "", "optional, pass the CA certificate file")
	pflag.StringVar(&cfg.TLS.Cert, "cert", "", "optional, pass the client certificate file")
	pflag.StringVar(&cfg.TLS.Key, "key", "", "optional, pass the certificate key file")
	pflag.Parse()

	if cfg.Help {
		pflag.Usage()
		os.Exit(0)
	}

	if pflag.NFlag() <= 0 {
		cfg.IsInteractive = true
	}

	if cfg.NoColor {
		color.NoColor = true
	}

	return cfg
}
