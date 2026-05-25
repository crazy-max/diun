package mail

import (
	"context"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	email "github.com/wneessen/go-mail"
)

func TestSendSMTPRegression(t *testing.T) {
	tests := []struct {
		name          string
		localName     string
		advertiseAuth bool
		wantHelo      string
	}{
		{
			name:      "configured localName is used",
			localName: "mail.example.com",
			wantHelo:  "mail.example.com",
		},
		{
			name:          "empty localName falls back to localhost without forcing auth",
			localName:     "",
			advertiseAuth: true,
			wantHelo:      "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, done := startSMTPServer(t, tt.advertiseAuth)
			cfg := testMailConfig(t, addr)
			cfg.LocalName = tt.localName

			client := &Client{cfg: cfg}
			mailClient, err := client.mailClient("", "")
			require.NoError(t, err)

			require.NoError(t, mailClient.DialAndSend(testMessage(t)))
			result := waitSMTPServer(t, done)

			assert.Equal(t, tt.wantHelo, result.helo)
			if tt.advertiseAuth {
				assert.False(t, result.authSeen)
			}
		})
	}
}

func TestSendWrapsDeliveryError(t *testing.T) {
	addr := freeTCPAddr(t)
	cfg := testMailConfig(t, addr)
	cfg.TemplateTitle = "title"
	cfg.TemplateBody = "body"

	client := &Client{
		cfg: cfg,
		meta: model.Meta{
			Name: "Diun",
		},
	}

	err := client.Send(model.NotifEntry{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot send mail notification")
}

type smtpServerResult struct {
	err      error
	helo     string
	authSeen bool
}

func startSMTPServer(t *testing.T, advertiseAuth bool) (string, <-chan smtpServerResult) {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	done := make(chan smtpServerResult, 1)
	go func() {
		var result smtpServerResult
		defer func() {
			done <- result
		}()

		conn, err := listener.Accept()
		if err != nil {
			result.err = err
			return
		}
		defer conn.Close()

		tp := textproto.NewConn(conn)
		defer tp.Close()
		write := func(line string) {
			_ = tp.PrintfLine("%s", line)
		}

		write("220 localhost ESMTP")
		for {
			line, err := tp.ReadLine()
			if err != nil {
				result.err = err
				return
			}

			upperLine := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upperLine, "EHLO "), strings.HasPrefix(upperLine, "HELO "):
				_, helo, _ := strings.Cut(line, " ")
				result.helo = helo
				if advertiseAuth {
					write("250-localhost")
					write("250 AUTH PLAIN LOGIN")
					continue
				}
				write("250 localhost")
			case upperLine == "NOOP":
				write("250 OK")
			case strings.HasPrefix(upperLine, "AUTH "):
				result.authSeen = true
				write("535 authentication disabled")
			case strings.HasPrefix(upperLine, "MAIL FROM:"):
				write("250 OK")
			case strings.HasPrefix(upperLine, "RCPT TO:"):
				write("250 OK")
			case upperLine == "DATA":
				write("354 end data with <CR><LF>.<CR><LF>")
				_, err := tp.ReadDotLines()
				if err != nil {
					result.err = err
					return
				}
				write("250 OK")
			case upperLine == "RSET":
				write("250 OK")
			case upperLine == "QUIT":
				write("221 bye")
				return
			default:
				write("250 OK")
			}
		}
	}()

	t.Cleanup(func() {
		_ = listener.Close()
	})

	return listener.Addr().String(), done
}

func freeTCPAddr(t *testing.T) string {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())
	return addr
}

func waitSMTPServer(t *testing.T, done <-chan smtpServerResult) smtpServerResult {
	t.Helper()

	select {
	case result := <-done:
		require.NoError(t, result.err)
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for SMTP server")
		return smtpServerResult{}
	}
}

func testMailConfig(t *testing.T, addr string) *model.NotifMail {
	t.Helper()

	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	portNum, err := strconv.Atoi(port)
	require.NoError(t, err)

	cfg := (&model.NotifMail{}).GetDefaults()
	cfg.Host = host
	cfg.Port = portNum
	cfg.From = "diun@example.com"
	cfg.To = []string{"ops@example.com", "dev@example.com"}
	return cfg
}

func testMessage(t *testing.T) *email.Msg {
	t.Helper()

	message := email.NewMsg()
	require.NoError(t, message.From("diun@example.com"))
	require.NoError(t, message.To("ops@example.com"))
	message.Subject("test")
	message.SetBodyString(email.TypeTextPlain, "test")
	return message
}
