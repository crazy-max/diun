// SPDX-FileCopyrightText: Copyright (c) The go-mail Authors
//
// SPDX-License-Identifier: MIT

package smtp

import (
	"net/textproto"
	"strings"
	"sync"
)

// ResponseErrorHandler provides custom handling for SMTP response errors.
//
// This interface defines a method for handling SMTP responses that do not comply with expected
// formats or behaviors. It is useful for implementing retry logic, logging, or protocol-specific
// error handling logic. It injects itself into the smtp.Client and is called whenever a server
// response does fail.
//
// Parameters:
//   - host: The hostname of the SMTP server that this ResponseErrorHandler should handle
//   - command: The SMTP command that triggered the error.
//   - conn: The textproto.Conn to the SMTP server.
//   - err: The error received from the SMTP server.
//
// Returns:
//   - An error indicating the outcome of the error handling logic.
type ResponseErrorHandler interface {
	HandleError(host, command string, conn *textproto.Conn, err error) error
}

// DefaultErrorHandler implements ResponseErrorHandler by returning the original error.
//
// This handler provides a default implementation of the ResponseErrorHandler interface,
// where no custom error processing is performed. It simply returns the original error
// received from the SMTP server.
//
// Returns:
//   - The original error passed to the handler.
type DefaultErrorHandler struct{}

// HandleError satisfies the ResponseErrorHandler interface for the DefaultErrorHandler type
func (d *DefaultErrorHandler) HandleError(_, _ string, _ *textproto.Conn, err error) error {
	return err
}

// HandlerKey uniquely identifies a host-command pair for handler mapping.
//
// This struct is used to associate a specific SMTP host and command combination with
// a ResponseErrorHandler. It enables mapping and retrieval of handlers based on
// the source of the error.
type HandlerKey struct {
	Host    string
	Command string
}

// ErrorHandlerRegistry manages custom error handlers for SMTP host-command pairs.
//
// This struct stores mappings between HandlerKey values and corresponding
// ResponseErrorHandler implementations. It supports concurrent access and provides
// a fallback default handler when no specific match is found.
type ErrorHandlerRegistry struct {
	mu             sync.RWMutex
	handlers       map[HandlerKey]ResponseErrorHandler
	defaultHandler ResponseErrorHandler
}

// NewErrorHandlerRegistry creates a new ErrorHandlerRegistry instance.
//
// This function initializes an ErrorHandlerRegistry with an empty handler map and
// assigns a DefaultErrorHandler as the fallback handler.
//
// Returns:
//   - A pointer to the newly constructed ErrorHandlerRegistry.
func NewErrorHandlerRegistry() *ErrorHandlerRegistry {
	return &ErrorHandlerRegistry{
		handlers:       make(map[HandlerKey]ResponseErrorHandler),
		defaultHandler: &DefaultErrorHandler{}, // Set the default handler
	}
}

// RegisterHandler associates a custom handler with a specific host and command.
//
// This method registers a ResponseErrorHandler for the given SMTP host and command
// combination. It ensures thread-safe access to the internal handler map.
//
// Parameters:
//   - host: The SMTP server hostname.
//   - command: The SMTP command to associate with the handler.
//   - handler: The ResponseErrorHandler to register.
func (r *ErrorHandlerRegistry) RegisterHandler(host, command string, handler ResponseErrorHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[HandlerKey{Host: strings.ToLower(host), Command: strings.ToLower(command)}] = handler
}

// GetHandler retrieves the handler for a given host and command.
//
// This method returns the ResponseErrorHandler registered for the specified SMTP host
// and command. If no handler is found, it returns the default handler.
//
// Parameters:
//   - host: The SMTP server hostname.
//   - command: The SMTP command to look up.
//
// Returns:
//   - The corresponding ResponseErrorHandler, or the default handler if none is registered.
func (r *ErrorHandlerRegistry) GetHandler(host, command string) ResponseErrorHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if handler, ok := r.handlers[HandlerKey{Host: strings.ToLower(host), Command: strings.ToLower(command)}]; ok {
		return handler
	}
	return r.defaultHandler
}

// SetDefaultHandler overrides the default ResponseErrorHandler.
//
// This method sets a new default handler to be used when no specific handler is
// registered for a host and command combination. It ensures thread-safe access.
//
// Parameters:
//   - handler: The new default ResponseErrorHandler to assign.
func (r *ErrorHandlerRegistry) SetDefaultHandler(handler ResponseErrorHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultHandler = handler
}
