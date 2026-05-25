// SPDX-FileCopyrightText: The go-mail Authors
//
// SPDX-License-Identifier: MIT

//go:build !gomailnotpl

package mail

import (
	"bytes"
	"errors"
	"fmt"
	ht "html/template"
	tt "text/template"
)

// SetBodyTextTemplate sets the body of the message from a given text/template.Template pointer.
//
// This method sets the body of the message using the provided text template and data. The content type
// will be set to "text/plain" automatically. The method executes the template with the provided data
// and writes the output to the message body. If the template is nil or fails to execute, an error will
// be returned.
//
// Parameters:
//   - tpl: A pointer to the text/template.Template to be used for the message body.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the body part.
//
// Returns:
//   - An error if the template is nil or fails to execute, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) SetBodyTextTemplate(tpl *tt.Template, data interface{}, opts ...PartOption) error {
	if tpl == nil {
		return errors.New(errTplPointerNil)
	}
	buffer := bytes.NewBuffer(nil)
	if err := tpl.Execute(buffer, data); err != nil {
		return fmt.Errorf(errTplExecuteFailed, err)
	}
	writeFunc := writeFuncFromBuffer(buffer)
	m.SetBodyWriter(TypeTextPlain, writeFunc, opts...)
	return nil
}

// SetBodyHTMLTemplate sets the body of the message from a given html/template.Template pointer.
//
// This method sets the body of the message using the provided HTML template and data. The content type
// will be set to "text/html" automatically. The method executes the template with the provided data
// and writes the output to the message body. If the template is nil or fails to execute, an error will
// be returned.
//
// Parameters:
//   - tpl: A pointer to the html/template.Template to be used for the message body.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the body part.
//
// Returns:
//   - An error if the template is nil or fails to execute, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) SetBodyHTMLTemplate(tpl *ht.Template, data interface{}, opts ...PartOption) error {
	if tpl == nil {
		return errors.New(errTplPointerNil)
	}
	buffer := bytes.NewBuffer(nil)
	if err := tpl.Execute(buffer, data); err != nil {
		return fmt.Errorf(errTplExecuteFailed, err)
	}
	writeFunc := writeFuncFromBuffer(buffer)
	m.SetBodyWriter(TypeTextHTML, writeFunc, opts...)
	return nil
}

// AddAlternativeTextTemplate sets the alternative body of the message to a text/template.Template output.
//
// The content type will be set to "text/plain" automatically. This method executes the provided text template
// with the given data and adds the result as an alternative version of the message body. If the template
// is nil or fails to execute, an error will be returned.
//
// Parameters:
//   - tpl: A pointer to the text/template.Template to be used for the alternative body.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the alternative body part.
//
// Returns:
//   - An error if the template is nil or fails to execute, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) AddAlternativeTextTemplate(tpl *tt.Template, data interface{}, opts ...PartOption) error {
	if tpl == nil {
		return errors.New(errTplPointerNil)
	}
	buffer := bytes.NewBuffer(nil)
	if err := tpl.Execute(buffer, data); err != nil {
		return fmt.Errorf(errTplExecuteFailed, err)
	}
	writeFunc := writeFuncFromBuffer(buffer)
	m.AddAlternativeWriter(TypeTextPlain, writeFunc, opts...)
	return nil
}

// AddAlternativeHTMLTemplate sets the alternative body of the message to an html/template.Template output.
//
// The content type will be set to "text/html" automatically. This method executes the provided HTML template
// with the given data and adds the result as an alternative version of the message body. If the template
// is nil or fails to execute, an error will be returned.
//
// Parameters:
//   - tpl: A pointer to the html/template.Template to be used for the alternative body.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the alternative body part.
//
// Returns:
//   - An error if the template is nil or fails to execute, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) AddAlternativeHTMLTemplate(tpl *ht.Template, data interface{}, opts ...PartOption) error {
	if tpl == nil {
		return errors.New(errTplPointerNil)
	}
	buffer := bytes.NewBuffer(nil)
	if err := tpl.Execute(buffer, data); err != nil {
		return fmt.Errorf(errTplExecuteFailed, err)
	}
	writeFunc := writeFuncFromBuffer(buffer)
	m.AddAlternativeWriter(TypeTextHTML, writeFunc, opts...)
	return nil
}

// AttachTextTemplate adds the output of a text/template.Template pointer as a File attachment to the Msg.
//
// This method allows you to attach the rendered output of a text template as a file to the message.
// The template is executed with the provided data, and its output is attached as a file. If the template
// fails to execute, an error will be returned.
//
// Parameters:
//   - name: The name of the file to be attached.
//   - tpl: A pointer to the text/template.Template to be executed for the attachment.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the attachment.
//
// Returns:
//   - An error if the template fails to execute or cannot be attached, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) AttachTextTemplate(
	name string, tpl *tt.Template, data interface{}, opts ...FileOption,
) error {
	file, err := fileFromTextTemplate(name, tpl, data)
	if err != nil {
		return fmt.Errorf("failed to attach template: %w", err)
	}
	m.attachments = m.appendFile(m.attachments, file, opts...)
	return nil
}

// AttachHTMLTemplate adds the output of a html/template.Template pointer as a File attachment to the Msg.
//
// This method allows you to attach the rendered output of an HTML template as a file to the message.
// The template is executed with the provided data, and its output is attached as a file. If the template
// fails to execute, an error will be returned.
//
// Parameters:
//   - name: The name of the file to be attached.
//   - tpl: A pointer to the html/template.Template to be executed for the attachment.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the attachment.
//
// Returns:
//   - An error if the template fails to execute or cannot be attached, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) AttachHTMLTemplate(
	name string, tpl *ht.Template, data interface{}, opts ...FileOption,
) error {
	file, err := fileFromHTMLTemplate(name, tpl, data)
	if err != nil {
		return fmt.Errorf("failed to attach template: %w", err)
	}
	m.attachments = m.appendFile(m.attachments, file, opts...)
	return nil
}

// EmbedTextTemplate adds the output of a text/template.Template pointer as an embedded File to the Msg.
//
// This method embeds the rendered output of a text template into the email message. The template is
// executed with the provided data, and its output is embedded as a file. If the template fails to execute,
// an error will be returned.
//
// Parameters:
//   - name: The name of the embedded file.
//   - tpl: A pointer to the text/template.Template to be executed for the embedded content.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the embedded file.
//
// Returns:
//   - An error if the template fails to execute or cannot be embedded, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) EmbedTextTemplate(
	name string, tpl *tt.Template, data interface{}, opts ...FileOption,
) error {
	file, err := fileFromTextTemplate(name, tpl, data)
	if err != nil {
		return fmt.Errorf("failed to embed template: %w", err)
	}
	m.embeds = m.appendFile(m.embeds, file, opts...)
	return nil
}

// EmbedHTMLTemplate adds the output of a html/template.Template pointer as an embedded File to the Msg.
//
// This method embeds the rendered output of an HTML template into the email message. The template is
// executed with the provided data, and its output is embedded as a file. If the template fails to execute,
// an error will be returned.
//
// Parameters:
//   - name: The name of the embedded file.
//   - tpl: A pointer to the html/template.Template to be executed for the embedded content.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the embedded file.
//
// Returns:
//   - An error if the template fails to execute or cannot be embedded, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) EmbedHTMLTemplate(
	name string, tpl *ht.Template, data interface{}, opts ...FileOption,
) error {
	file, err := fileFromHTMLTemplate(name, tpl, data)
	if err != nil {
		return fmt.Errorf("failed to embed template: %w", err)
	}
	m.embeds = m.appendFile(m.embeds, file, opts...)
	return nil
}

// fileFromTextTemplate returns a File pointer from a given text/template.Template.
//
// This method executes the provided text template with the given data and creates a File structure
// representing the output. The rendered template content is stored in a buffer and then processed
// as a file attachment or embed.
//
// Parameters:
//   - name: The name of the file to be created from the template output.
//   - tpl: A pointer to the text/template.Template to be executed.
//   - data: The data to populate the template.
//
// Returns:
//   - A pointer to the File structure representing the rendered template.
//   - An error if the template is nil or if it fails to execute.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func fileFromTextTemplate(name string, tpl *tt.Template, data interface{}) (*File, error) {
	if tpl == nil {
		return nil, errors.New(errTplPointerNil)
	}
	buffer := bytes.Buffer{}
	if err := tpl.Execute(&buffer, data); err != nil {
		return nil, fmt.Errorf(errTplExecuteFailed, err)
	}
	return fileFromReader(name, &buffer)
}

// fileFromHTMLTemplate returns a File pointer from a given html/template.Template.
//
// This method executes the provided HTML template with the given data and creates a File structure
// representing the output. The rendered template content is stored in a buffer and then processed
// as a file attachment or embed.
//
// Parameters:
//   - name: The name of the file to be created from the template output.
//   - tpl: A pointer to the html/template.Template to be executed.
//   - data: The data to populate the template.
//
// Returns:
//   - A pointer to the File structure representing the rendered template.
//   - An error if the template is nil or if it fails to execute.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func fileFromHTMLTemplate(name string, tpl *ht.Template, data interface{}) (*File, error) {
	if tpl == nil {
		return nil, errors.New(errTplPointerNil)
	}
	buffer := bytes.Buffer{}
	if err := tpl.Execute(&buffer, data); err != nil {
		return nil, fmt.Errorf(errTplExecuteFailed, err)
	}
	return fileFromReader(name, &buffer)
}
