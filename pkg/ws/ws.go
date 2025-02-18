package ws

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

func Connect(cfg config.Config) (*websocket.Conn, func(), error) {

	closeFunc := func() {}

	if cfg.ConnectURL == "" {
		return nil, closeFunc, fmt.Errorf("connect url is empty")
	}

	u, err := url.Parse(cfg.ConnectURL)
	if err != nil {
		return nil, closeFunc, fmt.Errorf("error while passing the url : %w", err)
	}

	headers := http.Header{}
	for _, h := range cfg.Headers {
		headSpl := strings.Split(h, ":")
		if len(headSpl) != 2 {
			return nil, closeFunc, fmt.Errorf("invalid header : %s", h)
		}
		headers.Set(headSpl[0], headSpl[1])
	}

	if cfg.Origin != "" {
		headers.Set("Origin", cfg.Origin)
	}

	if cfg.Auth != "" {
		headers.Set("Authorization", basicAuth(cfg.Auth))
	}

	dialer := websocket.Dialer{
		Subprotocols:    cfg.SubProtocol,
		TLSClientConfig: getTLSConfig(cfg.TLS, cfg.NoCertificateCheck),
	}

	if cfg.Proxy != "" {
		proxyURLParsed, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, closeFunc, fmt.Errorf("error while parsing the proxy url : %w", err)
		}
		dialer.Proxy = http.ProxyURL(proxyURLParsed)
	}

	c, resp, err := dialer.Dial(u.String(), headers)
	if err != nil {
		return nil, closeFunc, fmt.Errorf("dial error : %w", err)
	}

	if cfg.Response {
		for k, v := range resp.Header {
			log.Println(k, v)
		}

	}

	closeFunc = func() {
		if err := c.Close(); err != nil {
			logger.GlobalLogger.Debug().Err(err).Msg("error while closing the connection")
		}
	}

	return c, closeFunc, nil
}

func basicAuth(auth string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

var BlueColor = color.New(color.FgBlue).SprintfFunc()
var GreenColor = color.New(color.FgGreen).SprintfFunc()

func ReadMessages(cfg config.Config, conn *websocket.Conn, wg *sync.WaitGroup, l *readline.Instance) {

	defer wg.Done()

	fn := func(what string) func(appData string) error {
		return func(appData string) error {
			if cfg.ShowPingPong {
				log.Println(BlueColor("received %s (data: %s)", what, appData))
			}
			return nil
		}
	}

	conn.SetPingHandler(fn("ping"))
	conn.SetPongHandler(fn("pong"))

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			logger.GlobalLogger.Err(err).Msg("read error")
			l.Close()
			return
		}

		switch mt {
		case websocket.TextMessage:
			log.Println(GreenColor("« %s", string(message)))
		case websocket.BinaryMessage:
			log.Println(hex.EncodeToString(message))
		case websocket.CloseMessage:
			log.Println("received close message", message)
			return
		}

	}

}

func WriteToServer(conn *websocket.Conn, message string) {

	if conn == nil {
		logger.GlobalLogger.Error().Msg("Connection is nil")
	} else {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			logger.GlobalLogger.Err(err).Msg("write error")
		}
	}

}

func getTLSConfig(cfg config.TLS, noCheck bool) *tls.Config {

	if noCheck {
		return &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	caCertPool := processCACert(cfg.CA)

	certificates, err := processCert(cfg.Cert, cfg.Key, cfg.Passphrase)
	if err != nil {
		logger.GlobalLogger.Fatal().Err(err).Msg("error while processing client certificate")
		return nil
	}

	tlsCfg := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: certificates,
	}

	return tlsCfg
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
		logger.GlobalLogger.Fatal().Err(err).Msg("error while reading CA certificate")
		return nil
	}
	caCertPool := x509.NewCertPool()

	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		logger.GlobalLogger.Fatal().Err(err).Msg("error while parsing CA certificate")
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
