package ws

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/global"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

type CloseFunc func()

type ReaderFunc func(*websocket.Conn)

func Connect() (*websocket.Conn, CloseFunc, ReaderFunc, error) {

	closeFunc := func() {}
	rFunc := ReaderFunc(func(*websocket.Conn) {})

	if config.Flags.ConnectURL == "" {
		return nil, closeFunc, rFunc, fmt.Errorf("connect url is empty")
	}

	connectURL := strings.Replace(config.Flags.ConnectURL, "%", "%25", 1)

	u, err := url.Parse(connectURL)
	if err != nil {
		return nil, closeFunc, rFunc, fmt.Errorf("error while passing the url : %w", err)
	}

	headers := http.Header{}
	for _, h := range config.Flags.Headers {
		headSpl := strings.Split(h, ":")
		if len(headSpl) != 2 {
			return nil, closeFunc, rFunc, fmt.Errorf("invalid header : %s", h)
		}
		headers.Set(headSpl[0], headSpl[1])
	}

	if config.Flags.Origin != "" {
		headers.Set("Origin", config.Flags.Origin)
	}

	if config.Flags.Auth != "" {
		headers.Set("Authorization", BasicAuth(config.Flags.Auth))
	}

	dialer := websocket.Dialer{
		Subprotocols:    config.Flags.SubProtocol,
		TLSClientConfig: GetTLSConfig(),
	}

	if config.Flags.Proxy != "" {
		proxyURLParsed, err := url.Parse(config.Flags.Proxy)
		if err != nil {
			return nil, closeFunc, rFunc, fmt.Errorf("error while parsing the proxy url : %w", err)
		}
		dialer.Proxy = http.ProxyURL(proxyURLParsed)
	}

	if config.Flags.BindAddress != "" || config.Flags.IPVersion != "" {
		network := "tcp"
		switch config.Flags.IPVersion {
		case "4":
			network = "tcp4"
		case "6":
			network = "tcp6"
		case "":
		default:
			return nil, closeFunc, rFunc, fmt.Errorf("invalid ip-version: %s. Use 4 or 6", config.Flags.IPVersion)
		}

		netDialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}

		if config.Flags.BindAddress != "" {
			addrWithPort := net.JoinHostPort(config.Flags.BindAddress, "0")
			localAddr, err := net.ResolveTCPAddr(network, addrWithPort)
			if err != nil {
				return nil, closeFunc, rFunc, fmt.Errorf("error resolving bind address: %w", err)
			}
			netDialer.LocalAddr = localAddr
		}

		dialer.NetDialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
			return netDialer.DialContext(ctx, network, addr)
		}
	}

	c, resp, err := dialer.Dial(u.String(), headers)
	if err != nil {
		return nil, closeFunc, rFunc, fmt.Errorf("dial error : %w", err)
	}

	if config.Flags.ShouldShowResponseHeaders {
		for k, v := range resp.Header {
			log.Println(k, v)
		}
	}

	closeFunc = func() {
		if err := c.Close(); err != nil {
			logger.Debug().Err(err).Msg("error while closing the connection")
		}
	}

	go PingWorker(c)

	return c, closeFunc, readMessages, nil
}

func PingWorker(c *websocket.Conn) {
	for range time.Tick(config.Flags.PingInterval) {
		err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(3*time.Second))
		if err != nil {
			if err.Error() == "websocket: close sent" {
				return
			}
			logger.Debug().Err(err).Msg("error while pinging")
		}
	}
}

func BasicAuth(auth string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

var BlueColor = color.New(color.FgBlue).SprintfFunc()
var GreenColor = color.New(color.FgGreen).SprintfFunc()

func readMessages(conn *websocket.Conn) {

	fn := func(what string) func(appData string) error {
		return func(appData string) error {
			if config.Flags.ShowPingPong {
				log.Println(BlueColor("received %s (data: %s)", what, appData))
			}
			return nil
		}
	}

	defer func() {
		logger.Debug().Msg("enabling global stop application flag")
		global.Stop()
	}()

	conn.SetPingHandler(fn("ping"))
	conn.SetPongHandler(fn("pong"))

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			log.Println(err.Error())
			return
		}

		switch mt {
		case websocket.TextMessage:
			log.Println(formatMessage(message))
		case websocket.BinaryMessage:
			if config.Flags.IsGzipResponse {
				gzBytes, err := unzipGzipBytes(message)
				if err != nil {
					logger.Err(err).Msg("error while unzipping bytes")
				} else {
					log.Println(gzBytes)
				}
			} else {
				log.Println(hex.EncodeToString(message))
			}
		case websocket.CloseMessage:
			log.Println("received close message", message)
			return
		}

	}

}

func unzipGzipBytes(gzipBytes []byte) (string, error) {
	reader := bytes.NewReader(gzipBytes)
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	unzippedBytes, err := io.ReadAll(gzipReader)
	if err != nil {
		return "", fmt.Errorf("failed to read unzipped data: %w", err)
	}

	return string(unzippedBytes), nil
}

func formatMessage(message []byte) string {

	if !config.Flags.IsJSONPrettyPrint {
		return GreenColor("« %s", message)
	}

	m := map[string]any{}
	if err := json.Unmarshal(message, &m); err != nil {
		logger.Debug().Err(err).Msg("UNMARSHAL ERR")
		return GreenColor("« %s", message)
	}

	jenc, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		logger.Debug().Err(err).Msg("MARSHALINDENT ERR")
		return GreenColor("« %s", message)
	}

	return GreenColor("%s", jenc)
}

func WriteToServer(conn *websocket.Conn, mt int, message []byte) {

	if conn == nil {
		logger.Error().Msg("Connection is nil")
		return
	}

	if !config.Flags.IsBinary {
		if err := conn.WriteMessage(mt, message); err != nil {
			logger.Err(err).Msg("write error")
		}
		return
	}

	dec, err := hex.DecodeString(string(message))
	if err != nil {
		logger.Err(err).Msg("error while doing decode string")
		return
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, dec); err != nil {
		logger.Err(err).Msg("write error")
	}

}

func GetTLSConfig() *tls.Config {

	if config.Flags.NoCertificateCheck {
		return &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	caCertPool := processCACert(config.Flags.TLS.CA)

	certificates, err := processCert(config.Flags.TLS.Cert, config.Flags.TLS.Key, config.Flags.TLS.Passphrase)
	if err != nil {
		logger.Fatal().Err(err).Msg("error while processing client certificate")
		return nil
	}

	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: certificates,
	}
}

func processCert(certPath, keyPath, passphrase string) ([]tls.Certificate, error) {

	if certPath == "" && keyPath == "" {
		return nil, nil
	}

	if certPath != "" && keyPath == "" {
		return nil, fmt.Errorf("key is required if certificate is provided")
	}

	cert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("error while reading certificate : %w", err)
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("error while reading key : %w", err)
	}

	dkey, err := decryptPrivateKey(key, []byte(passphrase))
	if err != nil {
		return nil, fmt.Errorf("error while decrypting private key with passphrase : %w", err)
	}

	clientCert, err := tls.X509KeyPair(cert, dkey)
	if err != nil {
		return nil, fmt.Errorf("error while creating client certificate : %w", err)
	}

	return []tls.Certificate{clientCert}, nil

}

func processCACert(caCertPath string) *x509.CertPool {

	if caCertPath == "" {
		return nil
	}

	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("error while reading CA certificate")
		return nil
	}
	caCertPool := x509.NewCertPool()

	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		logger.Fatal().Err(err).Msg("error while parsing CA certificate")
		return nil
	}

	return caCertPool

}

func decryptPrivateKey(keyPEM []byte, passphrase []byte) ([]byte, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type == "RSA PRIVATE KEY" || block.Type == "PRIVATE KEY" {
		// Unencrypted RSA private key
		return block.Bytes, nil
	} else if block.Type == "ENCRYPTED PRIVATE KEY" {

		if len(block.Bytes) < 8 {
			return nil, fmt.Errorf("invalid encrypted private key")
		}
		if len(passphrase) == 0 {
			return nil, fmt.Errorf("passphrase required")
		}
		if len(passphrase) > 24 {
			passphrase = passphrase[:24]
		}

		c, err := des.NewTripleDESCipher(passphrase)
		if err != nil {
			return nil, err
		}
		iv := block.Bytes[:8]
		mode := cipher.NewCBCDecrypter(c, iv)
		plaintext := make([]byte, len(block.Bytes)-8)
		mode.CryptBlocks(plaintext, block.Bytes[8:])
		return plaintext, nil

	} else if block.Type == "AES-256-CBC ENCRYPTED PRIVATE KEY" || block.Type == "AES-128-CBC ENCRYPTED PRIVATE KEY" {
		// Modern encrypted private key format (AES)
		block, _ := pem.Decode(keyPEM) // Decode again to get the correct block after checking type

		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block")
		}

		if len(block.Bytes) < 16 {
			return nil, fmt.Errorf("invalid AES encrypted private key")
		}

		if len(passphrase) == 0 {
			return nil, fmt.Errorf("passphrase required")
		}

		c, err := aes.NewCipher(passphrase) // AES key size depends on your encryption (16, 24, or 32 bytes)
		if err != nil {
			return nil, err
		}

		iv := block.Bytes[:16] // Initialization vector
		mode := cipher.NewCBCDecrypter(c, iv)
		plaintext := make([]byte, len(block.Bytes)-16)
		mode.CryptBlocks(plaintext, block.Bytes[16:])
		return plaintext, nil
	} else {
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
}
