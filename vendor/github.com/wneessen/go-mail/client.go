// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wneessen/go-mail/log"
	"github.com/wneessen/go-mail/smtp"
)

const (
	// DefaultPort is the default connection port to the SMTP server.
	DefaultPort = 25

	// DefaultPortSSL is the default connection port for SSL/TLS to the SMTP server.
	DefaultPortSSL = 465

	// DefaultPortTLS is the default connection port for STARTTLS to the SMTP server.
	DefaultPortTLS = 587

	// DefaultTimeout is the default connection timeout.
	DefaultTimeout = time.Second * 15

	// DefaultTLSPolicy specifies the default TLS policy for connections.
	DefaultTLSPolicy = TLSMandatory

	// DefaultTLSMinVersion defines the minimum TLS version to be used for secure connections.
	// Nowadays TLS 1.2 is assumed be a sane default.
	DefaultTLSMinVersion = tls.VersionTLS12
)

const (
	// DSNMailReturnHeadersOnly requests that only the message headers of the mail message are returned in
	// a DSN (Delivery Status Notification).
	//
	// https://datatracker.ietf.org/doc/html/rfc1891#section-5.3
	DSNMailReturnHeadersOnly DSNMailReturnOption = "HDRS"

	// DSNMailReturnFull requests that the entire mail message is returned in any failed  DSN
	// (Delivery Status Notification) issued for this recipient.
	//
	// https://datatracker.ietf.org/doc/html/rfc1891/#section-5.3
	DSNMailReturnFull DSNMailReturnOption = "FULL"

	// DSNRcptNotifyNever indicates that no DSN (Delivery Status Notifications) should be sent for the
	// recipient under any condition.
	//
	// https://datatracker.ietf.org/doc/html/rfc1891/#section-5.1
	DSNRcptNotifyNever DSNRcptNotifyOption = "NEVER"

	// DSNRcptNotifySuccess indicates that the sender requests a DSN (Delivery Status Notification) if the
	// message is successfully delivered.
	//
	// https://datatracker.ietf.org/doc/html/rfc1891/#section-5.1
	DSNRcptNotifySuccess DSNRcptNotifyOption = "SUCCESS"

	// DSNRcptNotifyFailure requests that a DSN (Delivery Status Notification) is issued if delivery of
	// a message fails.
	//
	// https://datatracker.ietf.org/doc/html/rfc1891/#section-5.1
	DSNRcptNotifyFailure DSNRcptNotifyOption = "FAILURE"

	// DSNRcptNotifyDelay indicates the sender's willingness to receive "delayed" DSNs.
	//
	// Delayed DSNs may be issued if delivery of a message has been delayed for an unusual amount of time
	// (as determined by the MTA at which the message is delayed), but the final delivery status (whether
	// successful or failure) cannot be determined. The absence of the DELAY keyword in a NOTIFY parameter
	// requests that a "delayed" DSN NOT be issued under any conditions.
	//
	// https://datatracker.ietf.org/doc/html/rfc1891/#section-5.1
	DSNRcptNotifyDelay DSNRcptNotifyOption = "DELAY"
)

type (
	// DialContextFunc defines a function type for establishing a network connection using context, network
	// type, and address. It is used to specify custom DialContext function.
	//
	// By default we use net.Dial or tls.Dial respectively.
	DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

	// DSNMailReturnOption is a type wrapper for a string and specifies the type of return content requested
	// in a Delivery Status Notification (DSN).
	//
	// https://datatracker.ietf.org/doc/html/rfc1891/
	DSNMailReturnOption string

	// DSNRcptNotifyOption is a type wrapper for a string and specifies the notification options for a
	// recipient in DSNs.
	//
	// https://datatracker.ietf.org/doc/html/rfc1891/
	DSNRcptNotifyOption string

	// Option is a function type that modifies the configuration or behavior of a Client instance.
	Option func(*Client) error

	// Client is responsible for connecting and interacting with an SMTP server.
	//
	// This struct represents the go-mail client, which manages the connection, authentication, and communication
	// with an SMTP server. It contains various configuration options, including connection timeouts, encryption
	// settings, authentication methods, and Delivery Status Notification (DSN) preferences.
	//
	// References:
	//   - https://datatracker.ietf.org/doc/html/rfc3207#section-2
	//   - https://datatracker.ietf.org/doc/html/rfc8314
	Client struct {
		// ErrorHandlerRegistry provides access to the smtp.Client's custom error handlers for SMTP
		// host-command pairs which are based on the smtp.ResponseErrorHandler interface.
		//
		// The smtp.ResponseErrorHandler interface defines a method for handling SMTP responses that do not
		// comply with expected formats or behaviors. It is useful for implementing retry logic, logging,
		// or error handling logic for non-compliant SMTP responses.
		ErrorHandlerRegistry *smtp.ErrorHandlerRegistry

		// connTimeout specifies timeout for the connection to the SMTP server.
		connTimeout time.Duration

		// dialContextFunc is the DialContextFunc that is used by the Client to connect to the SMTP server.
		dialContextFunc DialContextFunc

		// dsnRcptNotifyType represents the different types of notifications for DSN (Delivery Status Notifications)
		// receipts.
		dsnRcptNotifyType []string

		// dsnReturnType specifies the type of Delivery Status Notification (DSN) that should be requested for an
		// email.
		dsnReturnType DSNMailReturnOption

		// fallbackPort is used as an alternative port number in case the primary port is unavailable or
		// fails to bind.
		//
		// The fallbackPort is only used in combination with SetTLSPortPolicy and SetSSLPort correspondingly.
		fallbackPort int

		// helo is the hostname used in the HELO/EHLO greeting, that is sent to the target SMTP server.
		//
		// helo might be different as host. This can be useful in a shared-hosting scenario.
		helo string

		// host is the hostname of the SMTP server we are connecting to.
		host string

		// logAuthData indicates whether authentication-related data should be logged.
		logAuthData bool

		// logger is a logger that satisfies the log.Logger interface.
		logger log.Logger

		// mutex is used to synchronize access to shared resources, ensuring that only one goroutine can
		// modify them at a time.
		mutex sync.RWMutex

		// noNoop indicates that the Client should skip the "NOOP" command during the dial.
		//
		// This is useful for servers which delay potentially unwanted clients when they perform commands
		// other than AUTH.
		noNoop bool

		// pass represents a password or a secret token used for the SMTP authentication.
		pass string

		// port specifies the network port that is used to establish the connection with the SMTP server.
		port int

		// requestDSN indicates wether we want to request DSN (Delivery Status Notifications).
		requestDSN bool

		// sendMutex is used to synchronize access to shared resources during the dial and send methods.
		sendMutex sync.Mutex

		// skipUTF8 indicates that the Client should skip the "SMTPUTF8" in a "MAIL FROM" even if the server
		// claims to support it
		skipUTF8 bool

		// smtpAuth is the authentication type that is used to authenticate the user with SMTP server. It
		// satisfies the smtp.Auth interface.
		//
		// Unless you plan to write you own custom authentication method, it is advised to not set this manually.
		// You should use one of go-mail's SMTPAuthType, instead.
		smtpAuth smtp.Auth

		// smtpAuthType specifies the authentication type to be used for SMTP authentication.
		smtpAuthType SMTPAuthType

		// smtpClient is an instance of smtp.Client used for handling the communication with the SMTP server.
		smtpClient *smtp.Client

		// tlspolicy defines the TLSPolicy configuration the Client uses for the STARTTLS protocol.
		//
		// https://datatracker.ietf.org/doc/html/rfc3207#section-2
		tlspolicy TLSPolicy

		// tlsconfig is a pointer to tls.Config that specifies the TLS configuration for the STARTTLS communication.
		tlsconfig *tls.Config

		// useDebugLog indicates whether debug level logging is enabled for the Client.
		useDebugLog bool

		// user represents a username used for the SMTP authentication.
		user string

		// useUnixSocket indicates that a connection is established via a Unix Domain Socket instead of TCP
		useUnixSocket bool

		// useSSL indicates whether to use SSL/TLS encryption for network communication.
		//
		// https://datatracker.ietf.org/doc/html/rfc8314
		useSSL bool
	}
)

var (
	// ErrInvalidPort is returned when the specified port for the SMTP connection is not valid
	ErrInvalidPort = errors.New("invalid port number")

	// ErrInvalidTimeout is returned when the specified timeout is zero or negative.
	ErrInvalidTimeout = errors.New("timeout cannot be zero or negative")

	// ErrInvalidHELO is returned when the HELO/EHLO value is invalid due to being empty.
	ErrInvalidHELO = errors.New("invalid HELO/EHLO value - must not be empty")

	// ErrInvalidTLSConfig is returned when the provided TLS configuration is invalid or nil.
	ErrInvalidTLSConfig = errors.New("invalid TLS config")

	// ErrNoHostname is returned when the hostname for the client is not provided or empty.
	ErrNoHostname = errors.New("hostname for client cannot be empty")

	// ErrDeadlineExtendFailed is returned when an attempt to extend the connection deadline fails.
	ErrDeadlineExtendFailed = errors.New("connection deadline extension failed")

	// ErrNoActiveConnection indicates that there is no active connection to the SMTP server.
	ErrNoActiveConnection = errors.New("not connected to SMTP server")

	// ErrServerNoUnencoded indicates that the server does not support 8BITMIME for unencoded 8-bit messages.
	ErrServerNoUnencoded = errors.New("message is 8bit unencoded, but server does not support 8BITMIME")

	// ErrInvalidDSNMailReturnOption is returned when an invalid DSNMailReturnOption is provided as argument
	// to the WithDSN Option.
	ErrInvalidDSNMailReturnOption = errors.New("DSN mail return option can only be HDRS or FULL")

	// ErrInvalidDSNRcptNotifyOption is returned when an invalid DSNRcptNotifyOption is provided as argument
	// to the WithDSN Option.
	ErrInvalidDSNRcptNotifyOption = errors.New("DSN rcpt notify option can only be: NEVER, " +
		"SUCCESS, FAILURE or DELAY")

	// ErrInvalidDSNRcptNotifyCombination is returned when an invalid combination of DSNRcptNotifyOption is
	// provided as argument to the WithDSN Option.
	ErrInvalidDSNRcptNotifyCombination = errors.New("DSN rcpt notify option NEVER cannot be " +
		"combined with any of SUCCESS, FAILURE or DELAY")

	// ErrSMTPAuthMethodIsNil indicates that the SMTP authentication method provided is nil
	ErrSMTPAuthMethodIsNil = errors.New("SMTP auth method is nil")

	// ErrDialContextFuncIsNil indicates that a required dial context function is not provided.
	ErrDialContextFuncIsNil = errors.New("dial context function is nil")

	// ErrClientIsNil indicates that a required smtp client is not provided.
	ErrClientIsNil = errors.New("client is nil")
)

// NewClient creates a new Client instance with the provided host and optional configuration Option functions.
//
// This function initializes a Client with default values, such as connection timeout, port, TLS settings,
// and the HELO/EHLO hostname. Option functions, if provided, can override the default configuration.
// It ensures that essential values, like the host, are set. The function also supports connections to
// UNIX domain sockets by recognizing a "unix://" prefix in the host string and adjusting the configuration
// accordingly. An error is returned if critical defaults are unset.
//
// Parameters:
//   - host: The hostname of the SMTP server to connect to, or a UNIX domain socket prefixed with "unix://".
//   - opts: Optional configuration functions to override default settings.
//
// Returns:
//   - A pointer to the initialized Client.
//   - An error if any critical default values are missing or options fail to apply.
func NewClient(host string, opts ...Option) (*Client, error) {
	c := &Client{
		ErrorHandlerRegistry: smtp.NewErrorHandlerRegistry(),
		smtpAuthType:         SMTPAuthNoAuth,
		connTimeout:          DefaultTimeout,
		host:                 host,
		port:                 DefaultPort,
		tlsconfig:            &tls.Config{ServerName: host, MinVersion: DefaultTLSMinVersion},
		tlspolicy:            DefaultTLSPolicy,
	}

	// Set default HELO/EHLO hostname
	if err := c.setDefaultHelo(); err != nil {
		return c, err
	}

	// Override defaults with optionally provided Option functions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			return c, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// We allow connecting to a UNIX Domain Socket
	if strings.HasPrefix(c.host, "unix://") {
		c.useUnixSocket = true
		c.host = strings.TrimPrefix(c.host, "unix://")
	}

	// Some settings in a Client cannot be empty/unset
	if c.host == "" {
		return c, ErrNoHostname
	}

	return c, nil
}

// WithPort sets the port number for the Client and overrides the default port.
//
// This function sets the specified port number for the Client, ensuring that the port number is valid
// (between 1 and 65535). If the provided port number is invalid, an error is returned.
//
// Parameters:
//   - port: The port number to be used by the Client. Must be between 1 and 65535.
//
// Returns:
//   - An Option function that applies the port setting to the Client.
//   - An error if the port number is outside the valid range.
func WithPort(port int) Option {
	return func(c *Client) error {
		if port < 1 || port > 65535 {
			return ErrInvalidPort
		}
		c.port = port
		return nil
	}
}

// WithTimeout sets the connection timeout for the Client and overrides the default timeout.
//
// This function configures the Client with a specified connection timeout duration. It validates that the
// provided timeout is greater than zero. If the timeout is invalid, an error is returned.
//
// Parameters:
//   - timeout: The duration to be set as the connection timeout. Must be greater than zero.
//
// Returns:
//   - An Option function that applies the timeout setting to the Client.
//   - An error if the timeout duration is invalid.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		if timeout <= 0 {
			return ErrInvalidTimeout
		}
		c.connTimeout = timeout
		return nil
	}
}

// WithSSL enables implicit SSL/TLS for the Client.
//
// This function configures the Client to use implicit SSL/TLS for secure communication.
//
// Returns:
//   - An Option function that enables SSL/TLS for the Client.
func WithSSL() Option {
	return func(c *Client) error {
		c.useSSL = true
		return nil
	}
}

// WithSSLPort enables implicit SSL/TLS with an optional fallback for the Client. The correct port is
// automatically set.
//
// When this option is used with NewClient, the default port 25 is overridden with port 465 for SSL/TLS connections.
// If fallback is set to true and the SSL/TLS connection fails, the Client attempts to connect on port 25 using an
// unencrypted connection. If WithPort has already been used to set a different port, that port takes precedence,
// and the automatic fallback mechanism is skipped.
//
// Parameters:
//   - fallback: A boolean indicating whether to fall back to port 25 without SSL/TLS if the connection fails.
//
// Returns:
//   - An Option function that enables SSL/TLS and configures the fallback mechanism for the Client.
func WithSSLPort(fallback bool) Option {
	return func(c *Client) error {
		c.SetSSLPort(true, fallback)
		return nil
	}
}

// WithDebugLog enables debug logging for the Client.
//
// This function activates debug logging, which logs incoming and outgoing communication between the
// Client and the SMTP server to os.Stderr. By default the debug logging will redact any kind of SMTP
// authentication data. If you need access to the actual authentication data in your logs, you can
// enable authentication data logging with the WithLogAuthData option or by setting it with the
// Client.SetLogAuthData method.
//
// Returns:
//   - An Option function that enables debug logging for the Client.
func WithDebugLog() Option {
	return func(c *Client) error {
		c.useDebugLog = true
		return nil
	}
}

// WithLogger defines a custom logger for the Client.
//
// This function sets a custom logger for the Client, which must satisfy the log.Logger interface. The custom
// logger is used only when debug logging is enabled. By default, log.Stdlog is used if no custom logger is provided.
//
// Parameters:
//   - logger: A logger that satisfies the log.Logger interface.
//
// Returns:
//   - An Option function that sets the custom logger for the Client.
func WithLogger(logger log.Logger) Option {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}

// WithHELO sets the HELO/EHLO string used by the Client.
//
// This function configures the HELO/EHLO string sent by the Client when initiating communication
// with the SMTP server. By default, os.Hostname is used to identify the HELO/EHLO string.
//
// Parameters:
//   - helo: The string to be used for the HELO/EHLO greeting. Must not be empty.
//
// Returns:
//   - An Option function that sets the HELO/EHLO string for the Client.
//   - An error if the provided HELO string is empty.
func WithHELO(helo string) Option {
	return func(c *Client) error {
		if helo == "" {
			return ErrInvalidHELO
		}
		c.helo = helo
		return nil
	}
}

// WithTLSPolicy sets the TLSPolicy of the Client and overrides the DefaultTLSPolicy.
//
// This function configures the Client's TLSPolicy, specifying how the Client handles TLS for SMTP connections.
// It overrides the default policy. For best practices regarding SMTP TLS connections, it is recommended to use
// WithTLSPortPolicy instead.
//
// Parameters:
//   - policy: The TLSPolicy to be applied to the Client.
//
// Returns:
//   - An Option function that sets the TLSPolicy for the Client.
func WithTLSPolicy(policy TLSPolicy) Option {
	return func(c *Client) error {
		c.tlspolicy = policy
		return nil
	}
}

// WithTLSPortPolicy enables explicit TLS via STARTTLS for the Client using the provided TLSPolicy. The
// correct port is automatically set.
//
// When TLSMandatory or TLSOpportunistic is provided as the TLSPolicy, port 587 is used for the connection.
// If the connection fails with TLSOpportunistic, the Client attempts to connect on port 25 using an unencrypted
// connection as a fallback. If NoTLS is specified, the Client will always use port 25.
// If WithPort has already been used to set a different port, that port takes precedence, and the automatic fallback
// mechanism is skipped.
//
// Parameters:
//   - policy: The TLSPolicy to be used for STARTTLS communication.
//
// Returns:
//   - An Option function that sets the TLSPortPolicy for the Client.
func WithTLSPortPolicy(policy TLSPolicy) Option {
	return func(c *Client) error {
		c.SetTLSPortPolicy(policy)
		return nil
	}
}

// WithTLSConfig sets the tls.Config for the Client and overrides the default configuration.
//
// This function configures the Client with a custom tls.Config. It overrides the default TLS settings.
// An error is returned if the provided tls.Config is nil or invalid.
//
// Parameters:
//   - tlsconfig: A pointer to a tls.Config struct to be used for the Client. Must not be nil.
//
// Returns:
//   - An Option function that sets the tls.Config for the Client.
//   - An error if the provided tls.Config is invalid.
func WithTLSConfig(tlsconfig *tls.Config) Option {
	return func(c *Client) error {
		if tlsconfig == nil {
			return ErrInvalidTLSConfig
		}
		c.tlsconfig = tlsconfig
		return nil
	}
}

// WithSMTPAuth configures the Client to use the specified SMTPAuthType for SMTP authentication.
//
// This function sets the Client to use the specified SMTPAuthType for authenticating with the SMTP server.
//
// Parameters:
//   - authtype: The SMTPAuthType to be used for SMTP authentication.
//
// Returns:
//   - An Option function that configures the Client to use the specified SMTPAuthType.
func WithSMTPAuth(authtype SMTPAuthType) Option {
	return func(c *Client) error {
		c.smtpAuthType = authtype
		return nil
	}
}

// WithSMTPAuthCustom sets a custom SMTP authentication mechanism for the Client.
//
// This function configures the Client to use a custom SMTP authentication mechanism. The provided
// mechanism must satisfy the smtp.Auth interface.
//
// Parameters:
//   - smtpAuth: The custom SMTP authentication mechanism, which must implement the smtp.Auth interface.
//
// Returns:
//   - An Option function that sets the custom SMTP authentication for the Client.
func WithSMTPAuthCustom(smtpAuth smtp.Auth) Option {
	return func(c *Client) error {
		if smtpAuth == nil {
			return ErrSMTPAuthMethodIsNil
		}
		c.smtpAuth = smtpAuth
		c.smtpAuthType = SMTPAuthCustom
		return nil
	}
}

// WithUsername sets the username that the Client will use for SMTP authentication.
//
// This function configures the Client with the specified username for SMTP authentication.
//
// Important:
//   - Specifying a username with this option alone does NOT enable SMTP authentication.
//   - To actually perform authentication with the server, you must also configure an
//     authentication mechanism by using either WithSMTPAuth() or WithSMTPAuthCustom().
//   - If you only call WithUsername() without setting an SMTP authentication method,
//     the provided username will be stored but never used.
//
// Parameters:
//   - username: The username to be used for SMTP authentication.
//
// Returns:
//   - An Option function that sets the username for the Client.
func WithUsername(username string) Option {
	return func(c *Client) error {
		c.user = username
		return nil
	}
}

// WithPassword sets the password that the Client will use for SMTP authentication.
//
// This function configures the Client with the specified password for SMTP authentication.
//
// Important:
//   - Specifying a password with this option alone does NOT enable SMTP authentication.
//   - To actually perform authentication with the server, you must also configure an
//     authentication mechanism by using either WithSMTPAuth() or WithSMTPAuthCustom().
//   - If you only call WithPassword() without setting an SMTP authentication method,
//     the provided password will be stored but never used.
//
// Parameters:
//   - password: The password to be used for SMTP authentication.
//
// Returns:
//   - An Option function that sets the password for the Client.
func WithPassword(password string) Option {
	return func(c *Client) error {
		c.pass = password
		return nil
	}
}

// WithDSN enables DSN (Delivery Status Notifications) for the Client as described in RFC 1891.
//
// This function configures the Client to request DSN, which provides status notifications for email delivery.
// DSN is only effective if the SMTP server supports it. By default, DSNMailReturnOption is set to DSNMailReturnFull,
// and DSNRcptNotifyOption is set to DSNRcptNotifySuccess and DSNRcptNotifyFailure.
//
// Returns:
//   - An Option function that enables DSN for the Client.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc1891
func WithDSN() Option {
	return func(c *Client) error {
		c.requestDSN = true
		c.dsnReturnType = DSNMailReturnFull
		c.dsnRcptNotifyType = []string{string(DSNRcptNotifyFailure), string(DSNRcptNotifySuccess)}
		return nil
	}
}

// WithDSNMailReturnType enables DSN (Delivery Status Notifications) for the Client as described in RFC 1891.
//
// This function configures the Client to request DSN and sets the DSNMailReturnOption to the provided value.
// DSN is only effective if the SMTP server supports it. The provided option must be either DSNMailReturnHeadersOnly
// or DSNMailReturnFull; otherwise, an error is returned.
//
// Parameters:
//   - option: The DSNMailReturnOption to be used (DSNMailReturnHeadersOnly or DSNMailReturnFull).
//
// Returns:
//   - An Option function that sets the DSNMailReturnOption for the Client.
//   - An error if an invalid DSNMailReturnOption is provided.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc1891
func WithDSNMailReturnType(option DSNMailReturnOption) Option {
	return func(c *Client) error {
		switch option {
		case DSNMailReturnHeadersOnly:
		case DSNMailReturnFull:
		default:
			return ErrInvalidDSNMailReturnOption
		}

		c.requestDSN = true
		c.dsnReturnType = option
		return nil
	}
}

// WithDSNRcptNotifyType enables DSN (Delivery Status Notifications) for the Client as described in RFC 1891.
//
// This function configures the Client to request DSN and sets the DSNRcptNotifyOption to the provided values.
// The provided options must be valid DSNRcptNotifyOption types. If DSNRcptNotifyNever is combined with
// any other notification type (such as DSNRcptNotifySuccess, DSNRcptNotifyFailure, or DSNRcptNotifyDelay),
// an error is returned.
//
// Parameters:
//   - opts: A variadic list of DSNRcptNotifyOption values (e.g., DSNRcptNotifySuccess, DSNRcptNotifyFailure).
//
// Returns:
//   - An Option function that sets the DSNRcptNotifyOption for the Client.
//   - An error if invalid DSNRcptNotifyOption values are provided or incompatible combinations are used.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc1891
func WithDSNRcptNotifyType(opts ...DSNRcptNotifyOption) Option {
	return func(c *Client) error {
		var rcptOpts []string
		var ns, nns bool
		if len(opts) > 0 {
			for _, opt := range opts {
				switch opt {
				case DSNRcptNotifyNever:
					ns = true
				case DSNRcptNotifySuccess:
					nns = true
				case DSNRcptNotifyFailure:
					nns = true
				case DSNRcptNotifyDelay:
					nns = true
				default:
					return ErrInvalidDSNRcptNotifyOption
				}
				rcptOpts = append(rcptOpts, string(opt))
			}
		}
		if ns && nns {
			return ErrInvalidDSNRcptNotifyCombination
		}

		c.requestDSN = true
		c.dsnRcptNotifyType = rcptOpts
		return nil
	}
}

// WithoutNoop indicates that the Client should skip the "NOOP" command during the dial.
//
// This option is useful for servers that delay potentially unwanted clients when they perform
// commands other than AUTH, such as Microsoft's Exchange Tarpit.
//
// Returns:
//   - An Option function that configures the Client to skip the "NOOP" command.
func WithoutNoop() Option {
	return func(c *Client) error {
		c.noNoop = true
		return nil
	}
}

// WithDialContextFunc sets the provided DialContextFunc as the DialContext for connecting to the SMTP server.
//
// This function overrides the default DialContext function used by the Client when establishing a connection
// to the SMTP server with the provided DialContextFunc.
//
// Parameters:
//   - dialCtxFunc: The custom DialContextFunc to be used for connecting to the SMTP server.
//
// Returns:
//   - An Option function that sets the custom DialContextFunc for the Client.
func WithDialContextFunc(dialCtxFunc DialContextFunc) Option {
	return func(c *Client) error {
		if dialCtxFunc == nil {
			return ErrDialContextFuncIsNil
		}
		c.dialContextFunc = dialCtxFunc
		return nil
	}
}

// WithLogAuthData enables logging of authentication data.
//
// This function sets the logAuthData field of the Client to true, enabling the logging of authentication data.
//
// Be cautious when using this option, as the logs may include unencrypted authentication data, depending on
// the SMTP authentication method in use, which could pose a data protection risk.
//
// Returns:
//   - An Option function that configures the Client to enable authentication data logging.
func WithLogAuthData() Option {
	return func(c *Client) error {
		c.logAuthData = true
		return nil
	}
}

// WithoutSMTPUTF8 forces the SMTP client to skip the SMTPUTF8 extension in the "MAIL FROM" command, even if
// the server supports it.
//
// This option is useful for servers that advertise support for SMTPUTF8 but do not actually implement it.
//
// Returns:
//   - An Option function that configures the Client to skip SMTPUTF8 in the "MAIL FROM" command.
func WithoutSMTPUTF8() Option {
	return func(c *Client) error {
		c.skipUTF8 = true
		return nil
	}
}

// TLSPolicy returns the TLSPolicy that is currently set on the Client as a string.
//
// This method retrieves the current TLSPolicy configured for the Client and returns it as a string representation.
//
// Returns:
//   - A string representing the currently set TLSPolicy for the Client.
func (c *Client) TLSPolicy() string {
	return c.tlspolicy.String()
}

// ServerAddr returns the server address that is currently set on the Client in the format "host:port".
//
// This method constructs and returns the server address using the host and port currently configured
// for the Client.
//
// Returns:
//   - A string representing the server address in the format "host:port".
func (c *Client) ServerAddr() string {
	if c.useUnixSocket {
		return c.host
	}
	return fmt.Sprintf("%s:%d", c.host, c.port)
}

// SetTLSPolicy sets or overrides the TLSPolicy currently configured on the Client with the given TLSPolicy.
//
// This method allows the user to set a new TLSPolicy for the Client. For best practices regarding
// SMTP TLS connections, it is recommended to use SetTLSPortPolicy instead.
//
// Parameters:
//   - policy: The TLSPolicy to be set for the Client.
func (c *Client) SetTLSPolicy(policy TLSPolicy) {
	c.tlspolicy = policy
}

// SetTLSPortPolicy sets or overrides the TLSPolicy currently configured on the Client with the given TLSPolicy.
// The correct port is automatically set based on the specified policy.
//
// If TLSMandatory or TLSOpportunistic is provided as the TLSPolicy, port 587 will be used for the connection.
// If the connection fails with TLSOpportunistic, the Client will attempt to connect on port 25 using
// an unencrypted connection as a fallback. If NoTLS is provided, the Client will always use port 25.
//
// Note: If a different port has already been set using WithPort, that port takes precedence and is used
// to establish the SSL/TLS connection, skipping the automatic fallback mechanism.
//
// Parameters:
//   - policy: The TLSPolicy to be set for the Client.
func (c *Client) SetTLSPortPolicy(policy TLSPolicy) {
	if c.port == DefaultPort {
		c.port = DefaultPortTLS
		c.fallbackPort = 0

		if policy == TLSOpportunistic {
			c.fallbackPort = DefaultPort
		}
		if policy == NoTLS {
			c.port = DefaultPort
		}
	}

	c.tlspolicy = policy
}

// SetSSL sets or overrides whether the Client should use implicit SSL/TLS.
//
// This method configures the Client to either enable or disable implicit SSL/TLS for secure communication.
//
// Parameters:
//   - ssl: A boolean value indicating whether to enable (true) or disable (false) implicit SSL/TLS.
func (c *Client) SetSSL(ssl bool) {
	c.useSSL = ssl
}

// SetSSLPort sets or overrides whether the Client should use implicit SSL/TLS with optional fallback.
// The correct port is automatically set.
//
// If ssl is set to true, the default port 25 will be overridden with port 465. If fallback is set to true
// and the SSL/TLS connection fails, the Client will attempt to connect on port 25 using an unencrypted
// connection.
//
// Note: If a different port has already been set using WithPort, that port takes precedence and is used
// to establish the SSL/TLS connection, skipping the automatic fallback mechanism.
//
// Parameters:
//   - ssl: A boolean value indicating whether to enable implicit SSL/TLS.
//   - fallback: A boolean value indicating whether to enable fallback to an unencrypted connection.
func (c *Client) SetSSLPort(ssl bool, fallback bool) {
	if c.port == DefaultPort {
		if ssl {
			c.port = DefaultPortSSL
		}

		c.fallbackPort = 0
		if fallback {
			c.fallbackPort = DefaultPort
		}
	}

	c.useSSL = ssl
}

// SetDebugLog sets or overrides whether the Client is using debug logging. The debug logger will log incoming
// and outgoing communication between the client and the server to log.Logger that is defined on the Client.
//
// Note: The SMTP communication might include unencrypted authentication data, depending on whether you are using
// SMTP authentication and the type of authentication mechanism. This could pose a data protection risk. Use
// debug logging with caution.
//
// Parameters:
//   - val: A boolean value indicating whether to enable (true) or disable (false) debug logging.
func (c *Client) SetDebugLog(val bool) {
	c.SetDebugLogWithSMTPClient(c.smtpClient, val)
}

// SetDebugLogWithSMTPClient sets or overrides whether the provided smtp.Client is using debug logging.
// The debug logger will log incoming and outgoing communication between the client and the server to
// log.Logger that is defined on the Client.
//
// Note: The SMTP communication might include unencrypted authentication data, depending on whether you are using
// SMTP authentication and the type of authentication mechanism. This could pose a data protection risk. Use
// debug logging with caution.
//
// Parameters:
//   - client: A pointer to the smtp.Client that handles the connection to the server.
//   - val: A boolean value indicating whether to enable (true) or disable (false) debug logging.
func (c *Client) SetDebugLogWithSMTPClient(client *smtp.Client, val bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.useDebugLog = val
	if client != nil {
		client.SetDebugLog(val)
	}
}

// SetLogger sets or overrides the custom logger currently used by the Client. The logger must
// satisfy the log.Logger interface and is only utilized when debug logging is enabled on the
// Client.
//
// By default, log.Stdlog is used if no custom logger is provided.
//
// Parameters:
//   - logger: A logger that satisfies the log.Logger interface to be set for the Client.
func (c *Client) SetLogger(logger log.Logger) {
	c.SetLoggerWithSMTPClient(c.smtpClient, logger)
}

// SetLoggerWithSMTPClient sets or overrides the custom logger currently used by the provided smtp.Client.
// The logger must satisfy the log.Logger interface and is only utilized when debug logging is enabled on
// the provided smtp.Client.
//
// By default, log.Stdlog is used if no custom logger is provided.
//
// Parameters:
//   - client: A pointer to the smtp.Client that handles the connection to the server.
//   - logger: A logger that satisfies the log.Logger interface to be set for the Client.
func (c *Client) SetLoggerWithSMTPClient(client *smtp.Client, logger log.Logger) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.logger = logger
	if client != nil {
		client.SetLogger(logger)
	}
}

// SetTLSConfig sets or overrides the tls.Config currently configured for the Client with the
// given value. An error is returned if the provided tls.Config is invalid.
//
// This method ensures that the provided tls.Config is not nil before updating the Client's
// TLS configuration.
//
// Parameters:
//   - tlsconfig: A pointer to the tls.Config struct to be set for the Client. Must not be nil.
//
// Returns:
//   - An error if the provided tls.Config is invalid or nil.
func (c *Client) SetTLSConfig(tlsconfig *tls.Config) error {
	if tlsconfig == nil {
		return ErrInvalidTLSConfig
	}
	c.tlsconfig = tlsconfig
	return nil
}

// SetUsername sets or overrides the username that the Client will use for SMTP authentication.
//
// This method updates the username used by the Client for authenticating with the SMTP server.
//
// Parameters:
//   - username: The username to be set for SMTP authentication.
func (c *Client) SetUsername(username string) {
	c.user = username
}

// SetPassword sets or overrides the password that the Client will use for SMTP authentication.
//
// This method updates the password used by the Client for authenticating with the SMTP server.
//
// Parameters:
//   - password: The password to be set for SMTP authentication.
func (c *Client) SetPassword(password string) {
	c.pass = password
}

// SetSMTPAuth sets or overrides the SMTPAuthType currently configured on the Client for SMTP
// authentication.
//
// This method updates the authentication type used by the Client for authenticating with the
// SMTP server and resets any custom SMTP authentication mechanism.
//
// Parameters:
//   - authtype: The SMTPAuthType to be set for the Client.
func (c *Client) SetSMTPAuth(authtype SMTPAuthType) {
	c.smtpAuthType = authtype
	c.smtpAuth = nil
}

// SetSMTPAuthCustom sets or overrides the custom SMTP authentication mechanism currently
// configured for the Client. The provided authentication mechanism must satisfy the
// smtp.Auth interface.
//
// This method updates the authentication mechanism used by the Client for authenticating
// with the SMTP server and sets the authentication type to SMTPAuthCustom.
//
// Parameters:
//   - smtpAuth: The custom SMTP authentication mechanism to be set for the Client.
func (c *Client) SetSMTPAuthCustom(smtpAuth smtp.Auth) {
	c.smtpAuth = smtpAuth
	c.smtpAuthType = SMTPAuthCustom
}

// SetLogAuthData sets or overrides the logging of SMTP authentication data for the Client.
//
// This function sets the logAuthData field of the Client to true, enabling the logging of authentication data.
//
// Be cautious when using this option, as the logs may include unencrypted authentication data, depending on
// the SMTP authentication method in use, which could pose a data protection risk.
//
// Parameters:
//   - logAuth: Set wether or not to log SMTP authentication data for the Client.
func (c *Client) SetLogAuthData(logAuth bool) {
	c.logAuthData = logAuth
}

// DialWithContext establishes a connection to the server using the provided context.Context.
//
// This function adds a deadline based on the Client's timeout to the provided context.Context
// before connecting to the server. After dialing the defined DialContextFunc and successfully
// establishing the connection, it sends the HELO/EHLO SMTP command, followed by optional
// STARTTLS and SMTP AUTH commands. If debug logging is enabled, it attaches the log.Logger.
//
// After this method is called, the Client will have an active (cancelable) connection to the
// SMTP server.
//
// Parameters:
//   - ctxDial: The context.Context used to control the connection timeout and cancellation.
//
// Returns:
//   - An error if the connection to the SMTP server fails or any subsequent command fails.
func (c *Client) DialWithContext(ctxDial context.Context) error {
	client, err := c.DialToSMTPClientWithContext(ctxDial)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	c.smtpClient = client
	c.mutex.Unlock()
	return nil
}

// DialToSMTPClientWithContext establishes and configures a smtp.Client connection using
// the provided context.
//
// This function uses the provided context to manage the connection deadline and cancellation.
// It dials the SMTP server using the Client's configured DialContextFunc or a default dialer.
// If SSL is enabled, it uses a TLS connection. After successfully connecting, it initializes
// an smtp.Client, sends the HELO/EHLO command, and optionally performs STARTTLS and SMTP AUTH
// based on the Client's configuration. Debug and authentication logging are enabled if
// configured.
//
// Parameters:
//   - ctxDial: The context used to control the connection timeout and cancellation.
//
// Returns:
//   - A pointer to the initialized smtp.Client.
//   - An error if the connection fails, the smtp.Client cannot be created, or any subsequent commands fail.
func (c *Client) DialToSMTPClientWithContext(ctxDial context.Context) (*smtp.Client, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ctx, cancel := context.WithDeadline(ctxDial, time.Now().Add(c.connTimeout))
	defer cancel()

	isEncrypted := false
	dialContextFunc := c.dialContextFunc
	if c.dialContextFunc == nil {
		netDialer := net.Dialer{}
		dialContextFunc = netDialer.DialContext
		if c.useSSL {
			tlsDialer := tls.Dialer{NetDialer: &netDialer, Config: c.tlsconfig}
			isEncrypted = true
			dialContextFunc = tlsDialer.DialContext
		}
	}

	network := "tcp"
	if c.useUnixSocket {
		network = "unix"
	}

	connection, err := dialContextFunc(ctx, network, c.ServerAddr())
	if err != nil && !c.useUnixSocket && c.fallbackPort != 0 {
		// TODO: should we somehow log or append the previous error?
		connection, err = dialContextFunc(ctx, "tcp", c.serverFallbackAddr())
	}
	if err != nil {
		return nil, err
	}

	err = connection.SetDeadline(time.Now().Add(c.connTimeout))
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(connection, c.host)
	if err != nil {
		return nil, err
	}
	client.ErrorHandlerRegistry = c.ErrorHandlerRegistry

	err = client.UpdateDeadline(c.connTimeout)
	if err != nil {
		return nil, err
	}

	if c.logger != nil {
		client.SetLogger(c.logger)
	}
	if c.useDebugLog {
		client.SetDebugLog(true)
	}
	if c.logAuthData {
		client.SetLogAuthData()
	}
	client.SkipSMTPUTF8(c.skipUTF8)
	if err = client.Hello(c.helo); err != nil {
		return nil, err
	}

	if err = c.tls(client, &isEncrypted); err != nil {
		return nil, err
	}

	if err = c.auth(client, isEncrypted); err != nil {
		return nil, err
	}

	return client, nil
}

// Close terminates the connection to the SMTP server, returning an error if the disconnection
// fails. If the connection is already closed, this method is a no-op and disregards any error.
//
// This function checks if the Client's SMTP connection is active. If not, it simply returns
// without any action. If the connection is active, it attempts to gracefully close the
// connection using the Quit method.
//
// Returns:
//   - An error if the disconnection fails; otherwise, returns nil.
func (c *Client) Close() error {
	return c.CloseWithSMTPClient(c.smtpClient)
}

// CloseWithSMTPClient terminates the connection of the provided smtp.Client to the SMTP server,
// returning an error if the disconnection fails. If the connection is already closed, this
// method is a no-op and disregards any error.
//
// This function checks if the smtp.Client connection is active. If not, it simply returns
// without any action. If the connection is active, it attempts to gracefully close the
// connection using the Quit method.
//
// Parameters:
//   - client: A pointer to the smtp.Client that handles the connection to the server.
//
// Returns:
//   - An error if the disconnection fails; otherwise, returns nil.
func (c *Client) CloseWithSMTPClient(client *smtp.Client) error {
	if client == nil || !client.HasConnection() {
		return nil
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to close SMTP client: %w", err)
	}

	return nil
}

// Reset sends an SMTP RSET command to reset the state of the current SMTP session.
//
// This method checks the connection to the SMTP server and, if the connection is valid,
// it sends an RSET command to reset the session state. If the connection is invalid or
// the command fails, an error is returned.
//
// Returns:
//   - An error if the connection check fails or if sending the RSET command fails;
//     otherwise, returns nil.
func (c *Client) Reset() error {
	return c.ResetWithSMTPClient(c.smtpClient)
}

// ResetWithSMTPClient sends an SMTP RSET command to the provided smtp.Client, to reset
// the state of the current SMTP session.
//
// This method checks the connection to the SMTP server and, if the connection is valid,
// it sends an RSET command to reset the session state. If the connection is invalid or
// the command fails, an error is returned.
//
// Parameters:
//   - client: A pointer to the smtp.Client that handles the connection to the server.
//
// Returns:
//   - An error if the connection check fails or if sending the RSET command fails; otherwise, returns nil.
func (c *Client) ResetWithSMTPClient(client *smtp.Client) error {
	if err := c.checkConn(client); err != nil {
		return err
	}
	if err := client.Reset(); err != nil {
		return fmt.Errorf("failed to send RSET to SMTP client: %w", err)
	}

	return nil
}

// DialAndSend establishes a connection to the server and sends out the provided Msg.
// It calls DialAndSendWithContext with an empty Context.Background.
//
// This method simplifies the process of connecting to the SMTP server and sending messages
// by using a default context. It prepares the messages for sending and ensures the connection
// is established before attempting to send them.
//
// Parameters:
//   - messages: A variadic list of pointers to Msg objects to be sent.
//
// Returns:
//   - An error if the connection fails or if sending the messages fails; otherwise, returns nil.
func (c *Client) DialAndSend(messages ...*Msg) error {
	ctx := context.Background()
	return c.DialAndSendWithContext(ctx, messages...)
}

// DialAndSendWithContext establishes a connection to the SMTP server using DialWithContext
// with the provided context.Context, then sends out the given Msg. After successful delivery,
// the Client will close the connection to the server.
//
// This method first attempts to connect to the SMTP server using the provided context.
// Upon successful connection, it sends the specified messages and ensures that the connection
// is closed after the operation, regardless of success or failure in sending the messages.
//
// Parameters:
//   - ctx: The context.Context to control the connection timeout and cancellation.
//   - messages: A variadic list of pointers to Msg objects to be sent.
//
// Returns:
//   - An error if the connection fails, if sending the messages fails, or if closing the
//     connection fails; otherwise, returns nil.
func (c *Client) DialAndSendWithContext(ctx context.Context, messages ...*Msg) (err error) {
	client, err := c.DialToSMTPClientWithContext(ctx)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer func() {
		if closeErr := c.CloseWithSMTPClient(client); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close connection: %w", closeErr))
		}
	}()

	if err = c.SendWithSMTPClient(client, messages...); err != nil {
		return fmt.Errorf("send failed: %w", err)
	}

	return nil
}

// Send attempts to send one or more Msg using the SMTP client that is assigned to the Client.
// If the Client has no active connection to the server, Send will fail with an error. For
// each of the provided Msg, it will associate a SendError with the Msg in case of a
// transmission or delivery error.
//
// This method first checks for an active connection to the SMTP server. If the connection is
// not valid, it returns a SendError. It then iterates over the provided messages, attempting
// to send each one. If an error occurs during sending, the method records the error and
// associates it with the corresponding Msg. If multiple errors are encountered, it aggregates
// them into a single SendError to be returned.
//
// Parameters:
//   - client: A pointer to the smtp.Client that holds the connection to the SMTP server
//   - messages: A variadic list of pointers to Msg objects to be sent.
//
// Returns:
//   - An error that represents the sending result, which may include multiple SendErrors if
//     any occurred; otherwise, returns nil.
func (c *Client) Send(messages ...*Msg) (returnErr error) {
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()
	return c.SendWithSMTPClient(c.smtpClient, messages...)
}

// SendWithSMTPClient attempts to send one or more Msg using a provided smtp.Client with an
// established connection to the SMTP server. If the smtp.Client has no active connection to
// the server, SendWithSMTPClient will fail with an error. For each of the provided Msg, it
// will associate a SendError with the Msg in case of a transmission or delivery error.
//
// This method first checks for an active connection to the SMTP server. If the connection is
// not valid, it returns a SendError. It then iterates over the provided messages, attempting
// to send each one. If an error occurs during sending, the method records the error and
// associates it with the corresponding Msg. If multiple errors are encountered, it aggregates
// them into a single SendError to be returned.
//
// Parameters:
//   - client: A pointer to the smtp.Client that holds the connection to the SMTP server
//   - messages: A variadic list of pointers to Msg objects to be sent.
//
// Returns:
//   - An error that represents the sending result, which may include multiple SendErrors if
//     any occurred; otherwise, returns nil.
func (c *Client) SendWithSMTPClient(client *smtp.Client, messages ...*Msg) (returnErr error) {
	if client == nil {
		return &SendError{
			Reason: ErrConnCheck, errlist: []error{ErrClientIsNil}, isTemp: isTempError(ErrClientIsNil),
			errcode: errorCode(ErrClientIsNil), enhancedStatusCode: enhancedStatusCode(ErrClientIsNil, false),
		}
	}

	escSupport, _ := client.Extension("ENHANCEDSTATUSCODES")
	if err := c.checkConn(client); err != nil {
		return &SendError{
			Reason: ErrConnCheck, errlist: []error{err}, isTemp: isTempError(err),
			errcode: errorCode(err), enhancedStatusCode: enhancedStatusCode(err, escSupport),
		}
	}

	var errs []error
	defer func() {
		returnErr = errors.Join(errs...)
	}()

	for id, message := range messages {
		if message == nil {
			continue
		}
		if sendErr := c.sendSingleMsg(client, message); sendErr != nil {
			messages[id].sendError = sendErr
			errs = append(errs, sendErr)
		}
	}

	return returnErr
}

// auth attempts to authenticate the client using SMTP AUTH mechanisms. It checks the connection,
// determines the supported authentication methods, and applies the appropriate authentication
// type. An error is returned if authentication fails.
//
// By default NewClient sets the SMTP authentication type to SMTPAuthNoAuth, meaning, that no
// SMTP authentication will be performed. If the user makes use of SetSMTPAuth or initialzes the
// client with WithSMTPAuth, the SMTP authentication type will be set in the Client, forcing
// this method to determine if the server supports the selected authentication method and
// assigning the corresponding smtp.Auth function to it.
//
// If the user set a custom SMTP authentication function using SetSMTPAuthCustom or
// WithSMTPAuthCustom, we will not perform any detection and assignment logic and will trust
// the user with their provided smtp.Auth function.
//
// Finally, it attempts to authenticate the client using the selected method.
//
// Returns:
//   - An error if the connection check fails, if no supported authentication method is found,
//     or if the authentication process fails.
func (c *Client) auth(client *smtp.Client, isEnc bool) error {
	var smtpAuth smtp.Auth
	if c.smtpAuthType == SMTPAuthCustom {
		smtpAuth = c.smtpAuth
	}
	if c.smtpAuth == nil && c.smtpAuthType != SMTPAuthNoAuth {
		hasSMTPAuth, smtpAuthType := client.Extension("AUTH")
		if !hasSMTPAuth {
			return fmt.Errorf("server does not support SMTP AUTH")
		}

		authType := c.smtpAuthType
		if c.smtpAuthType == SMTPAuthAutoDiscover {
			discoveredType, err := c.authTypeAutoDiscover(smtpAuthType, isEnc)
			if err != nil {
				return err
			}
			authType = discoveredType
		}

		switch authType {
		case SMTPAuthPlain:
			if !strings.Contains(smtpAuthType, string(SMTPAuthPlain)) {
				return ErrPlainAuthNotSupported
			}
			smtpAuth = smtp.PlainAuth("", c.user, c.pass, c.host, false)
		case SMTPAuthPlainNoEnc:
			if !strings.Contains(smtpAuthType, string(SMTPAuthPlain)) {
				return ErrPlainAuthNotSupported
			}
			smtpAuth = smtp.PlainAuth("", c.user, c.pass, c.host, true)
		case SMTPAuthLogin:
			if !strings.Contains(smtpAuthType, string(SMTPAuthLogin)) {
				return ErrLoginAuthNotSupported
			}
			smtpAuth = smtp.LoginAuth(c.user, c.pass, c.host, false)
		case SMTPAuthLoginNoEnc:
			if !strings.Contains(smtpAuthType, string(SMTPAuthLogin)) {
				return ErrLoginAuthNotSupported
			}
			smtpAuth = smtp.LoginAuth(c.user, c.pass, c.host, true)
		case SMTPAuthCramMD5:
			if !strings.Contains(smtpAuthType, string(SMTPAuthCramMD5)) {
				return ErrCramMD5AuthNotSupported
			}
			smtpAuth = smtp.CRAMMD5Auth(c.user, c.pass)
		case SMTPAuthXOAUTH2:
			if !strings.Contains(smtpAuthType, string(SMTPAuthXOAUTH2)) {
				return ErrXOauth2AuthNotSupported
			}
			smtpAuth = smtp.XOAuth2Auth(c.user, c.pass)
		case SMTPAuthSCRAMSHA1:
			if !strings.Contains(smtpAuthType, string(SMTPAuthSCRAMSHA1)) {
				return ErrSCRAMSHA1AuthNotSupported
			}
			smtpAuth = smtp.ScramSHA1Auth(c.user, c.pass)
		case SMTPAuthSCRAMSHA256:
			if !strings.Contains(smtpAuthType, string(SMTPAuthSCRAMSHA256)) {
				return ErrSCRAMSHA256AuthNotSupported
			}
			smtpAuth = smtp.ScramSHA256Auth(c.user, c.pass)
		case SMTPAuthSCRAMSHA1PLUS:
			if !strings.Contains(smtpAuthType, string(SMTPAuthSCRAMSHA1PLUS)) {
				return ErrSCRAMSHA1PLUSAuthNotSupported
			}
			tlsConnState, err := client.GetTLSConnectionState()
			if err != nil {
				return err
			}
			smtpAuth = smtp.ScramSHA1PlusAuth(c.user, c.pass, tlsConnState)
		case SMTPAuthSCRAMSHA256PLUS:
			if !strings.Contains(smtpAuthType, string(SMTPAuthSCRAMSHA256PLUS)) {
				return ErrSCRAMSHA256PLUSAuthNotSupported
			}
			tlsConnState, err := client.GetTLSConnectionState()
			if err != nil {
				return err
			}
			smtpAuth = smtp.ScramSHA256PlusAuth(c.user, c.pass, tlsConnState)
		default:
			return fmt.Errorf("unsupported SMTP AUTH type %q", c.smtpAuthType)
		}
	}

	if smtpAuth != nil {
		if err := client.Auth(smtpAuth); err != nil {
			return fmt.Errorf("SMTP AUTH failed: %w", err)
		}
	}
	return nil
}

func (c *Client) authTypeAutoDiscover(supported string, isEnc bool) (SMTPAuthType, error) {
	if supported == "" {
		return "", ErrNoSupportedAuthDiscovered
	}
	preferList := []SMTPAuthType{
		SMTPAuthSCRAMSHA256PLUS, SMTPAuthSCRAMSHA256, SMTPAuthSCRAMSHA1PLUS, SMTPAuthSCRAMSHA1,
		SMTPAuthCramMD5, SMTPAuthPlain, SMTPAuthLogin,
	}
	if !isEnc {
		preferList = []SMTPAuthType{SMTPAuthSCRAMSHA256, SMTPAuthSCRAMSHA1, SMTPAuthCramMD5}
	}
	mechs := strings.Split(supported, " ")

	for _, item := range preferList {
		if sliceContains(mechs, string(item)) {
			return item, nil
		}
	}
	return "", ErrNoSupportedAuthDiscovered
}

func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// sendSingleMsg sends out a single message and returns an error if the transmission or
// delivery fails. It is invoked by the public Send methods.
//
// This method handles the process of sending a single email message through the SMTP
// client. It performs several checks and operations, including verifying the encoding,
// retrieving the sender and recipient addresses, and managing delivery status notifications
// (DSN). It attempts to send the message and handles any errors that occur during the
// transmission process, ensuring that any necessary cleanup is performed (such as resetting
// the SMTP client if an error occurs).
//
// Parameters:
//   - message: A pointer to the Msg object representing the email message to be sent.
//
// Returns:
//   - An error if any part of the sending process fails; otherwise, returns nil.
func (c *Client) sendSingleMsg(client *smtp.Client, message *Msg) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	escSupport, _ := client.Extension("ENHANCEDSTATUSCODES")

	if message.encoding == NoEncoding {
		if ok, _ := client.Extension("8BITMIME"); !ok {
			return &SendError{Reason: ErrNoUnencoded, isTemp: false, affectedMsg: message}
		}
	}
	from, err := message.GetSender(false)
	if err != nil {
		return &SendError{
			Reason: ErrGetSender, errlist: []error{err}, isTemp: isTempError(err),
			affectedMsg: message, errcode: errorCode(err),
			enhancedStatusCode: enhancedStatusCode(err, escSupport),
		}
	}
	rcpts, err := message.GetRecipients()
	if err != nil {
		return &SendError{
			Reason: ErrGetRcpts, errlist: []error{err}, isTemp: isTempError(err),
			affectedMsg: message, errcode: errorCode(err),
			enhancedStatusCode: enhancedStatusCode(err, escSupport),
		}
	}

	if c.requestDSN {
		if c.dsnReturnType != "" {
			client.SetDSNMailReturnOption(string(c.dsnReturnType))
		}
	}
	if err = client.Mail(from); err != nil {
		retError := &SendError{
			Reason: ErrSMTPMailFrom, errlist: []error{err}, isTemp: isTempError(err),
			affectedMsg: message, errcode: errorCode(err),
			enhancedStatusCode: enhancedStatusCode(err, escSupport),
		}
		if resetSendErr := client.Reset(); resetSendErr != nil {
			retError.errlist = append(retError.errlist, resetSendErr)
		}
		return retError
	}
	hasError := false
	rcptSendErr := &SendError{affectedMsg: message}
	rcptSendErr.errlist = make([]error, 0)
	rcptSendErr.rcpt = make([]string, 0)
	rcptNotifyOpt := strings.Join(c.dsnRcptNotifyType, ",")
	client.SetDSNRcptNotifyOption(rcptNotifyOpt)
	for _, rcpt := range rcpts {
		if err = client.Rcpt(rcpt); err != nil {
			rcptSendErr.Reason = ErrSMTPRcptTo
			rcptSendErr.errlist = append(rcptSendErr.errlist, err)
			rcptSendErr.rcpt = append(rcptSendErr.rcpt, rcpt)
			rcptSendErr.isTemp = isTempError(err)
			rcptSendErr.errcode = errorCode(err)
			rcptSendErr.enhancedStatusCode = enhancedStatusCode(err, escSupport)
			hasError = true
		}
	}
	if hasError {
		if resetSendErr := client.Reset(); resetSendErr != nil {
			rcptSendErr.errlist = append(rcptSendErr.errlist, resetSendErr)
		}
		return rcptSendErr
	}
	writer, err := client.Data()
	if err != nil {
		return &SendError{
			Reason: ErrSMTPData, errlist: []error{err}, isTemp: isTempError(err),
			affectedMsg: message, errcode: errorCode(err),
			enhancedStatusCode: enhancedStatusCode(err, escSupport),
		}
	}
	_, err = message.WriteTo(writer)
	if err != nil {
		return &SendError{
			Reason: ErrWriteContent, errlist: []error{err}, isTemp: isTempError(err),
			affectedMsg: message, errcode: errorCode(err),
			enhancedStatusCode: enhancedStatusCode(err, escSupport),
		}
	}
	if err = writer.Close(); err != nil {
		return &SendError{
			Reason: ErrSMTPDataClose, errlist: []error{err}, isTemp: isTempError(err),
			affectedMsg: message, errcode: errorCode(err),
			enhancedStatusCode: enhancedStatusCode(err, escSupport),
		}
	}
	if dc, ok := writer.(*smtp.DataCloser); ok {
		message.serverResponse = dc.ServerResponse()
	}
	message.isDelivered = true

	if err = c.ResetWithSMTPClient(client); err != nil {
		return &SendError{
			Reason: ErrSMTPReset, errlist: []error{err}, isTemp: isTempError(err),
			affectedMsg: message, errcode: errorCode(err),
			enhancedStatusCode: enhancedStatusCode(err, escSupport),
		}
	}
	return nil
}

// checkConn ensures that a required server connection is available and extends the connection
// deadline.
//
// This method verifies whether there is an active connection to the SMTP server. If there is no
// connection, it returns an error. If the "noNoop" flag is not set, it sends a NOOP command to
// the server to confirm the connection is still valid. Finally, it updates the connection
// deadline based on the specified timeout value. If any operation fails, the appropriate error
// is returned.
//
// Returns:
//   - An error if there is no active connection, if the NOOP command fails, or if extending
//     the deadline fails; otherwise, returns nil.
func (c *Client) checkConn(client *smtp.Client) error {
	if client == nil {
		return ErrNoActiveConnection
	}
	if !client.HasConnection() {
		return ErrNoActiveConnection
	}

	c.mutex.RLock()
	noNoop := c.noNoop
	c.mutex.RUnlock()
	if !noNoop {
		if err := client.Noop(); err != nil {
			return ErrNoActiveConnection
		}
	}

	if err := client.UpdateDeadline(c.connTimeout); err != nil {
		return ErrDeadlineExtendFailed
	}
	return nil
}

// serverFallbackAddr returns the currently set combination of hostname and fallback port.
//
// This method constructs and returns the server address using the host and fallback port
// currently configured for the Client. It is useful for establishing a connection when
// the primary port is unavailable.
//
// Returns:
//   - A string representing the server address in the format "host:fallbackPort".
func (c *Client) serverFallbackAddr() string {
	return fmt.Sprintf("%s:%d", c.host, c.fallbackPort)
}

// setDefaultHelo sets the HELO/EHLO hostname to the local machine's hostname.
//
// This method retrieves the local hostname using the operating system's hostname function
// and sets it as the HELO/EHLO string for the Client. If retrieving the hostname fails,
// an error is returned.
//
// Returns:
//   - An error if there is a failure in reading the local hostname; otherwise, returns nil.
func (c *Client) setDefaultHelo() error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to read local hostname: %w", err)
	}
	c.helo = hostname
	return nil
}

// tls establishes a TLS connection based on the client's TLS policy and configuration.
// Returns an error if no active connection exists or if a TLS error occurs.
//
// This method first checks if there is an active connection to the SMTP server. If SSL is not
// being used and the TLS policy is not set to NoTLS, it checks for STARTTLS support. Depending
// on the TLS policy (mandatory or opportunistic), it may initiate a TLS connection using the
// StartTLS method. The method also retrieves the TLS connection state to determine if the
// connection is encrypted and returns any errors encountered during these processes.
//
// Returns:
//   - An error if there is no active connection, if STARTTLS is required but not supported,
//     or if there are issues during the TLS handshake; otherwise, returns nil.
func (c *Client) tls(client *smtp.Client, isEnc *bool) error {
	if !c.useSSL && c.tlspolicy != NoTLS {
		hasStartTLS := false
		extension, _ := client.Extension("STARTTLS")
		if c.tlspolicy == TLSMandatory {
			hasStartTLS = true
			if !extension {
				return fmt.Errorf("STARTTLS mode set to: %q, but target host does not support STARTTLS",
					c.tlspolicy)
			}
		}
		if c.tlspolicy == TLSOpportunistic {
			if extension {
				hasStartTLS = true
			}
		}
		if hasStartTLS {
			if err := client.StartTLS(c.tlsconfig); err != nil {
				return err
			}
		}
		tlsConnState, err := client.GetTLSConnectionState()
		if err != nil {
			switch {
			case errors.Is(err, smtp.ErrNonTLSConnection):
				*isEnc = false
				return nil
			default:
				return fmt.Errorf("failed to get TLS connection state: %w", err)
			}
		}
		*isEnc = tlsConnState.HandshakeComplete
	}
	return nil
}
