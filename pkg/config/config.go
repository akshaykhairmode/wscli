package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

type Flag struct {
	ConnectURL          string
	BindAddress         string
	IPVersion           string
	Auth                string
	Headers             []string
	Origin              string
	Execute             []string
	Wait                time.Duration
	PrintOutputInterval time.Duration
	PingInterval        time.Duration
	SubProtocol         []string
	Proxy               string
	UnixSocket          string

	Perf Perf

	ShowPingPong              bool
	IsSlash                   bool
	NoCertificateCheck        bool
	Version                   bool
	Verbose                   bool
	NoColor                   bool
	ShouldShowResponseHeaders bool
	IsJSONPrettyPrint         bool
	IsBinary                  bool
	IsGzipResponse            bool
	IsPerf                    bool

	IsStdOut bool

	Help    bool
	IsSTDin bool // read from stdin, cannot send messages to the server other than what is in the stdin

	TLS TLS
}

type TLS struct {
	CA         string
	Cert       string
	Key        string
	Passphrase string
}

type Perf struct {
	TotalConns           uint          `yaml:"tc"`      //total connections which needs to be created.
	LoadMessage          string        `yaml:"lm"`      //the load message which needs to be sent to the server.
	MessageInterval      time.Duration `yaml:"mi"`      //at what interval we send the messages.
	AuthMessage          string        `yaml:"am"`      //the auth message which needs to be send as soon as connecting.
	WaitBeforeAuth       time.Duration `yaml:"wba"`     //wait for x amount of time before sending auth message
	WaitAfterAuth        time.Duration `yaml:"waa"`     //wait for x amount of time before starting to send load.
	RampUpConnsPerSecond uint          `yaml:"rups"`    //how many connections to add every second
	LogOutFile           string        `yaml:"outfile"` //give the file path where to write the logs
	ConfigPath           string        //the file path from where to get the perf config
}

func (p Perf) String() string {
	return fmt.Sprintf(`Total Connections: %d, Messages Interval: %s, Wait Before Auth: %s, Wait After Auth: %s
	Ramp Up Connections Per Second: %d, Log Out File: %s, Auth Message: %s, Load Message: %s
	ConfigPath : %s`,
		p.TotalConns,
		p.MessageInterval,
		p.WaitBeforeAuth,
		p.WaitAfterAuth,
		p.RampUpConnsPerSecond,
		p.LogOutFile,
		p.AuthMessage,
		p.LoadMessage,
		p.ConfigPath,
	)
}

var Flags *Flag

func init() {
	Flags = get()
}

func get() *Flag {
	cfg := Flag{}

	pflag.BoolVarP(&cfg.Help, "help", "h", false, "	Display help information.")
	pflag.BoolVar(&cfg.IsSlash, "slash", false, "Enable slash commands (Experimental).")
	pflag.BoolVarP(&cfg.NoCertificateCheck, "no-check", "n", false, "Disable TLS certificate verification.")
	pflag.BoolVarP(&cfg.ShowPingPong, "show-ping-pong", "P", false, "Show ping/pong messages.")
	pflag.BoolVarP(&cfg.Version, "version", "V", false, "Display version information.")
	pflag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable debug logging.")
	pflag.BoolVar(&cfg.NoColor, "no-color", false, "Disable colored output.")
	pflag.BoolVarP(&cfg.ShouldShowResponseHeaders, "response", "r", false, "Display HTTP response headers from the server.")
	pflag.BoolVar(&cfg.IsJSONPrettyPrint, "jspp", false, "Enable JSON pretty printing for responses.")
	pflag.BoolVarP(&cfg.IsBinary, "binary", "b", false, "Send hex encoded data to server")
	pflag.BoolVar(&cfg.IsGzipResponse, "gzipr", false, "Enable gzip decoding if server messages are gzip-encoded. (Note: Server must send messages as binary.)")
	pflag.BoolVar(&cfg.IsStdOut, "std-out", false, "print the received messages in standard output, default is standard error")

	pflag.StringVarP(&cfg.ConnectURL, "connect", "c", "", "WebSocket connection URL.")
	pflag.StringVar(&cfg.BindAddress, "bind-address", "", "Bind address for outgoing connection (e.g., 192.168.1.100).")
	pflag.StringVar(&cfg.IPVersion, "ip-version", "", "IP version to use for outgoing connection (4 or 6).")
	pflag.StringVar(&cfg.Proxy, "proxy", "", "Use a proxy URL.")
	pflag.StringVar(&cfg.UnixSocket, "unix-socket", "", "Connect to a Unix domain socket.")
	pflag.StringVar(&cfg.Auth, "auth", "", "HTTP Basic Authentication credentials (e.g., username:password).")
	pflag.StringSliceVarP(&cfg.Headers, "header", "H", []string{}, "Custom headers (key:value, can be used multiple times).")
	pflag.StringVarP(&cfg.Origin, "origin", "o", "", "Specify origin for the WebSocket connection (optional).")
	pflag.StringSliceVarP(&cfg.Execute, "execute", "x", []string{}, "Execute a command after connecting (use multiple times for multiple commands).")
	pflag.DurationVarP(&cfg.Wait, "wait", "w", 0, "Wait time after command execution (1s, 1m, 1h).")
	pflag.StringSliceVarP(&cfg.SubProtocol, "sub-protocol", "s", []string{}, "Specify a sub-protocol for the WebSocket connection (optional, can be used multiple times).")
	pflag.DurationVar(&cfg.PrintOutputInterval, "print-interval", time.Second, "how often to print the status on the terminal")
	pflag.DurationVar(&cfg.PingInterval, "ping-interval", 30*time.Second, "how often to ping the connections which are created")

	pflag.StringVar(&cfg.TLS.CA, "ca", "", "Path to the CA certificate file (optional).")
	pflag.StringVar(&cfg.TLS.Cert, "cert", "", "Path to the client certificate file (optional).")
	pflag.StringVar(&cfg.TLS.Key, "key", "", "Path to the certificate key file (optional).")

	//perf
	pflag.BoolVar(&cfg.IsPerf, "perf", false, "Enable load testing")
	pflag.StringVar(&cfg.Perf.ConfigPath, "pconfig", "", "Load perf config from file")
	pflag.UintVar(&cfg.Perf.TotalConns, "tc", 0, "Total number of connections to create")
	pflag.StringVar(&cfg.Perf.LoadMessage, "lm", "", "Load message to send to the server")
	pflag.DurationVar(&cfg.Perf.MessageInterval, "mi", 0, "the interval for sending messages.")
	pflag.StringVar(&cfg.Perf.AuthMessage, "am", "", "Authentication message to send to the server")
	pflag.DurationVar(&cfg.Perf.WaitAfterAuth, "waa", 0, "Wait time after authentication before sending load messages to server")
	pflag.DurationVar(&cfg.Perf.WaitBeforeAuth, "wba", 0, "Wait time before sending authentication to server")
	pflag.UintVar(&cfg.Perf.RampUpConnsPerSecond, "rups", 1, "Number of connections to ramp up per second")
	pflag.StringVar(&cfg.Perf.LogOutFile, "outfile", "", "Write to file instead of output on terminal")

	pflag.Parse()

	if cfg.Help {
		pflag.Usage()
		os.Exit(0)
	}

	if cfg.NoColor {
		color.NoColor = true
	}

	cfg.IsSTDin = isInputFromPipe()

	return &cfg
}

var IsSTDoutRedirected bool

func init() {
	fi, err := os.Stdout.Stat()
	if err != nil {
		IsSTDoutRedirected = false
		return
	}

	IsSTDoutRedirected = fi.Mode().IsRegular()
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
	sb.WriteString(fmt.Sprintf("  ConnectURL: %s\n", c.ConnectURL))
	sb.WriteString(fmt.Sprintf("  BindAddress: %s\n", c.BindAddress))
	sb.WriteString(fmt.Sprintf("  IPVersion: %s\n", c.IPVersion))
	sb.WriteString(fmt.Sprintf("  Auth: %s\n", c.Auth))
	sb.WriteString(fmt.Sprintf("  Headers: %v\n", c.Headers))
	sb.WriteString(fmt.Sprintf("  Origin: %s\n", c.Origin))
	sb.WriteString(fmt.Sprintf("  Execute: %v\n", c.Execute))
	sb.WriteString(fmt.Sprintf("  Wait: %s\n", c.Wait))
	sb.WriteString(fmt.Sprintf("  PrintOutputInterval: %s\n", c.PrintOutputInterval))
	sb.WriteString(fmt.Sprintf("  PingInterval: %s\n", c.PingInterval))
	sb.WriteString(fmt.Sprintf("  SubProtocol: %v\n", c.SubProtocol))
	sb.WriteString(fmt.Sprintf("  Proxy: %s\n", c.Proxy))

	sb.WriteString(fmt.Sprintf("  ShowPingPong: %t\n", c.ShowPingPong))
	sb.WriteString(fmt.Sprintf("  IsSlash: %t\n", c.IsSlash))
	sb.WriteString(fmt.Sprintf("  NoCertificateCheck: %t\n", c.NoCertificateCheck))
	sb.WriteString(fmt.Sprintf("  Version: %t\n", c.Version))
	sb.WriteString(fmt.Sprintf("  Verbose: %t\n", c.Verbose))
	sb.WriteString(fmt.Sprintf("  NoColor: %t\n", c.NoColor))
	sb.WriteString(fmt.Sprintf("  ShouldShowResponseHeaders: %t\n", c.ShouldShowResponseHeaders))
	sb.WriteString(fmt.Sprintf("  IsJSONPrettyPrint: %t\n", c.IsJSONPrettyPrint))
	sb.WriteString(fmt.Sprintf("  IsBinary: %t\n", c.IsBinary))
	sb.WriteString(fmt.Sprintf("  IsGzipResponse: %t\n", c.IsGzipResponse))
	sb.WriteString(fmt.Sprintf("  IsPerf: %t\n", c.IsPerf))
	sb.WriteString(fmt.Sprintf("  IsStdOut: %t\n", c.IsStdOut))

	sb.WriteString(fmt.Sprintf("  Help: %t\n", c.Help))
	sb.WriteString(fmt.Sprintf("  IsSTDin: %t\n", c.IsSTDin))

	sb.WriteString(fmt.Sprintf("  TLS: %+v\n", c.TLS))
	if c.IsPerf { // Added Perf details conditionally
		sb.WriteString("  Perf Config:\n")
		// Indent the Perf string output for better readability
		perfLines := strings.Split(c.Perf.String(), "\n")
		for _, line := range perfLines {
			sb.WriteString(fmt.Sprintf("    %s\n", strings.TrimSpace(line)))
		}
	}

	return sb.String()
}
