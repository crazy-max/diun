package mail

import (
	"context"
	"net"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	email "github.com/wneessen/go-mail"
)

func TestGenerateMessageID(t *testing.T) {
	// RFC 5322 Message-ID format: <local-part@domain>
	// local-part should contain: timestamp.randomhex
	messageIDRegex := regexp.MustCompile(`^<\d+\.[0-9a-f]+@.+>$`)

	tests := []struct {
		name       string
		cfg        *model.NotifMail
		wantDomain string
	}{
		{
			name: "with LocalName",
			cfg: &model.NotifMail{
				Host:      "smtp.example.com",
				LocalName: "mail.mydomain.com",
			},
			wantDomain: "mail.mydomain.com",
		},
		{
			name: "fallback to Host",
			cfg: &model.NotifMail{
				Host:      "smtp.example.com",
				LocalName: "",
			},
			wantDomain: "smtp.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				cfg: tt.cfg,
			}

			messageID, err := client.generateMessageID()
			require.NoError(t, err, "generateMessageID should not return an error")
			assert.NotEmpty(t, messageID, "Message-ID should not be empty")

			// Verify RFC 5322 format
			assert.True(t, messageIDRegex.MatchString(messageID),
				"Message-ID should match RFC 5322 format: %s", messageID)

			// Verify it contains the expected domain
			assert.Contains(t, messageID, tt.wantDomain,
				"Message-ID should contain domain %s: %s", tt.wantDomain, messageID)

			// Verify structure: <timestamp.randomhex@domain>
			assert.True(t, strings.HasPrefix(messageID, "<"), "Message-ID should start with <")
			assert.True(t, strings.HasSuffix(messageID, ">"), "Message-ID should end with >")
			assert.Contains(t, messageID, "@", "Message-ID should contain @")
			assert.Contains(t, messageID, ".", "Message-ID should contain . separator")
		})
	}
}

func TestGenerateMessageID_Uniqueness(t *testing.T) {
	client := &Client{
		cfg: &model.NotifMail{
			Host:      "smtp.example.com",
			LocalName: "mail.example.com",
		},
	}

	// Generate multiple Message-IDs and verify they are unique
	messageIDs := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		messageID, err := client.generateMessageID()
		require.NoError(t, err)
		assert.False(t, messageIDs[messageID], "Message-ID should be unique: %s", messageID)
		messageIDs[messageID] = true
	}

	assert.Len(t, messageIDs, iterations, "All generated Message-IDs should be unique")
}

func TestGenerateMessageID_Format(t *testing.T) {
	client := &Client{
		cfg: &model.NotifMail{
			Host:      "smtp.example.com",
			LocalName: "mail.example.com",
		},
	}

	messageID, err := client.generateMessageID()
	require.NoError(t, err)

	// Remove angle brackets
	messageID = strings.TrimPrefix(messageID, "<")
	messageID = strings.TrimSuffix(messageID, ">")

	// Split by @
	parts := strings.Split(messageID, "@")
	require.Len(t, parts, 2, "Message-ID should have exactly one @ symbol")

	localPart := parts[0]
	domain := parts[1]

	// Verify local part contains timestamp.randomhex
	localParts := strings.Split(localPart, ".")
	require.Len(t, localParts, 2, "Local part should be timestamp.randomhex")

	// Verify timestamp is numeric
	timestamp := localParts[0]
	assert.Regexp(t, `^\d+$`, timestamp, "Timestamp should be numeric")

	// Verify random part is hex (16 characters for 8 bytes)
	randomHex := localParts[1]
	assert.Len(t, randomHex, 16, "Random hex should be 16 characters (8 bytes)")
	assert.Regexp(t, `^[0-9a-f]+$`, randomHex, "Random part should be hex")

	// Verify domain
	assert.Equal(t, "mail.example.com", domain, "Domain should match LocalName")
}

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
