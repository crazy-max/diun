package mail

import (
	"regexp"
	"strings"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
