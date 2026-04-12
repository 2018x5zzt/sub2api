//go:build unit

package service

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEmailService_TestSMTPConnectionWithConfig_UsesStartTLSOnSubmissionPort(t *testing.T) {
	host, port, rootCAs, cleanup := startSMTPTestServer(t, smtpTestModeStartTLS)
	defer cleanup()

	restore := setSMTPTestRootCAs(rootCAs)
	defer restore()

	svc := NewEmailService(nil, nil)
	err := svc.TestSMTPConnectionWithConfig(&SMTPConfig{
		Host:     host,
		Port:     port,
		Username: "user@example.com",
		Password: "secret",
		UseTLS:   true,
	})
	require.NoError(t, err)
}

func TestEmailService_SendEmailWithConfig_UsesStartTLSOnSubmissionPort(t *testing.T) {
	host, port, rootCAs, cleanup := startSMTPTestServer(t, smtpTestModeStartTLS)
	defer cleanup()

	restore := setSMTPTestRootCAs(rootCAs)
	defer restore()

	svc := NewEmailService(nil, nil)
	err := svc.SendEmailWithConfig(&SMTPConfig{
		Host:     host,
		Port:     port,
		Username: "user@example.com",
		Password: "secret",
		From:     "user@example.com",
		FromName: "Sub2API",
		UseTLS:   true,
	}, "recipient@example.com", "SMTP Test", "<p>Hello</p>")
	require.NoError(t, err)
}

func TestEmailService_TestSMTPConnectionWithConfig_TimesOutWhenServerNeverGreets(t *testing.T) {
	host, port, cleanup := startBlackholeSMTPServer(t, time.Second)
	defer cleanup()

	prevTimeout := smtpDialTimeout
	smtpDialTimeout = 100 * time.Millisecond
	defer func() {
		smtpDialTimeout = prevTimeout
	}()

	svc := NewEmailService(nil, nil)
	startedAt := time.Now()
	err := svc.TestSMTPConnectionWithConfig(&SMTPConfig{
		Host:     host,
		Port:     port,
		Username: "user@example.com",
		Password: "secret",
		UseTLS:   false,
	})
	elapsed := time.Since(startedAt)

	require.Error(t, err)
	require.Less(t, elapsed, 500*time.Millisecond)
}

type smtpTestMode string

const (
	smtpTestModeStartTLS    smtpTestMode = "starttls"
	smtpTestModeImplicitTLS smtpTestMode = "implicit_tls"
)

func setSMTPTestRootCAs(rootCAs *x509.CertPool) func() {
	prev := smtpTLSRootCAs
	smtpTLSRootCAs = rootCAs
	return func() {
		smtpTLSRootCAs = prev
	}
}

func startSMTPTestServer(t *testing.T, mode smtpTestMode) (string, int, *x509.CertPool, func()) {
	t.Helper()

	tlsConfig, rootCAs := newSMTPTestTLSConfig(t)

	var (
		listener net.Listener
		err      error
	)
	if mode == smtpTestModeImplicitTLS {
		listener, err = tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	} else {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
	}
	require.NoError(t, err)

	var (
		wg        sync.WaitGroup
		serverErr error
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			if !isClosedNetworkErr(acceptErr) {
				serverErr = acceptErr
			}
			return
		}

		serveErr := serveSMTPTestConn(conn, tlsConfig, mode == smtpTestModeStartTLS)
		if serveErr != nil && !isClosedNetworkErr(serveErr) {
			serverErr = serveErr
		}
	}()

	host, portStr, splitErr := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, splitErr)
	port, convErr := strconv.Atoi(portStr)
	require.NoError(t, convErr)

	cleanup := func() {
		_ = listener.Close()
		wg.Wait()
		require.NoError(t, serverErr)
	}

	return host, port, rootCAs, cleanup
}

func startBlackholeSMTPServer(t *testing.T, holdFor time.Duration) (string, int, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		time.Sleep(holdFor)
	}()

	host, portStr, splitErr := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, splitErr)
	port, convErr := strconv.Atoi(portStr)
	require.NoError(t, convErr)

	cleanup := func() {
		_ = listener.Close()
		wg.Wait()
	}

	return host, port, cleanup
}

func serveSMTPTestConn(conn net.Conn, tlsConfig *tls.Config, allowStartTLS bool) error {
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	tlsActive := !allowStartTLS

	writeResponse := func(lines ...string) error {
		for _, line := range lines {
			if _, err := writer.WriteString(line + "\r\n"); err != nil {
				return err
			}
		}
		return writer.Flush()
	}

	readLine := func() (string, error) {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	if err := writeResponse("220 localhost ESMTP ready"); err != nil {
		return err
	}

	for {
		line, err := readLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO ") || strings.HasPrefix(upper, "HELO "):
			if allowStartTLS && !tlsActive {
				if err := writeResponse(
					"250-localhost",
					"250-STARTTLS",
					"250 AUTH PLAIN",
				); err != nil {
					return err
				}
			} else {
				if err := writeResponse(
					"250-localhost",
					"250 AUTH PLAIN",
				); err != nil {
					return err
				}
			}
		case upper == "STARTTLS":
			if !allowStartTLS || tlsActive {
				return fmt.Errorf("unexpected STARTTLS command")
			}
			if err := writeResponse("220 Ready to start TLS"); err != nil {
				return err
			}

			tlsConn := tls.Server(conn, tlsConfig)
			if err := tlsConn.Handshake(); err != nil {
				return err
			}

			conn = tlsConn
			reader = bufio.NewReader(conn)
			writer = bufio.NewWriter(conn)
			tlsActive = true
		case strings.HasPrefix(upper, "AUTH PLAIN"):
			if err := writeResponse("235 2.7.0 Authentication successful"); err != nil {
				return err
			}
		case strings.HasPrefix(upper, "MAIL FROM:"):
			if err := writeResponse("250 2.1.0 OK"); err != nil {
				return err
			}
		case strings.HasPrefix(upper, "RCPT TO:"):
			if err := writeResponse("250 2.1.5 OK"); err != nil {
				return err
			}
		case upper == "DATA":
			if err := writeResponse("354 End data with <CR><LF>.<CR><LF>"); err != nil {
				return err
			}
			for {
				dataLine, dataErr := readLine()
				if dataErr != nil {
					return dataErr
				}
				if dataLine == "." {
					break
				}
			}
			if err := writeResponse("250 2.0.0 Accepted"); err != nil {
				return err
			}
		case upper == "QUIT":
			_ = writeResponse("221 2.0.0 Bye")
			return nil
		default:
			return fmt.Errorf("unexpected SMTP command: %q", line)
		}
	}
}

func newSMTPTestTLSConfig(t *testing.T) (*tls.Config, *x509.CertPool) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "127.0.0.1",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	parsedCert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(parsedCert)

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}, rootCAs
}

func isClosedNetworkErr(err error) bool {
	return errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection")
}
