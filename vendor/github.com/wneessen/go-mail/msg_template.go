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
func (m *Msg) SetBodyTextTemplate(tpl *tt.Template, data any, opts ...PartOption) error {
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

// SetBodyNamedTextTemplate sets the body of the message from a template associated with a given
// text/template.Template pointer that has the given name.
//
// This method sets the body of the message using the provided text template, name and data. The content type
// will be set to "text/plain" automatically. The method executes the template with the provided data
// and writes the output to the message body. If the template is nil or fails to execute, an error will
// be returned.
//
// Parameters:
//   - tpl: A pointer to the text/template.Template containing a named template.
//   - name: Name of the template to be used for the message body.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the body part.
//
// Returns:
//   - An error if the template is nil or fails to execute, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) SetBodyNamedTextTemplate(tpl *tt.Template, name string, data any, opts ...PartOption) error {
	if tpl == nil {
		return errors.New(errTplPointerNil)
	}
	buffer := bytes.NewBuffer(nil)
	if err := tpl.ExecuteTemplate(buffer, name, data); err != nil {
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
func (m *Msg) SetBodyHTMLTemplate(tpl *ht.Template, data any, opts ...PartOption) error {
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

// SetBodyNamedHTMLTemplate sets the body of the message from a template associated with a given
// html/template.Template pointer that has the given name.
//
// This method sets the body of the message using the provided HTML template, name and data. The content type
// will be set to "text/html" automatically. The method executes the template with the provided data
// and writes the output to the message body. If the template is nil or fails to execute, an error will
// be returned.
//
// Parameters:
//   - tpl: A pointer to the html/template.Template containing a named template.
//   - name: Name of the template to be used for the message body.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the body part.
//
// Returns:
//   - An error if the template is nil or fails to execute, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) SetBodyNamedHTMLTemplate(tpl *ht.Template, name string, data any, opts ...PartOption) error {
	if tpl == nil {
		return errors.New(errTplPointerNil)
	}
	buffer := bytes.NewBuffer(nil)
	if err := tpl.ExecuteTemplate(buffer, name, data); err != nil {
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
func (m *Msg) AddAlternativeTextTemplate(tpl *tt.Template, data any, opts ...PartOption) error {
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

// AddAlternativeNamedTextTemplate sets the body of the message from a template associated with a given
// text/template.Template pointer that has the given name.
//
// The content type will be set to "text/plain" automatically. This method executes the named template provided by text template
// with the given data and adds the result as an alternative version of the message body. If the template
// is nil or fails to execute, an error will be returned.
//
// Parameters:
//   - tpl: A pointer to the text/template.Template containing a named template.
//   - name: Name of the template to be used for the alternative body.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the alternative body part.
//
// Returns:
//   - An error if the template is nil or fails to execute, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) AddAlternativeNamedTextTemplate(tpl *tt.Template, name string, data any, opts ...PartOption) error {
	if tpl == nil {
		return errors.New(errTplPointerNil)
	}
	buffer := bytes.NewBuffer(nil)
	if err := tpl.ExecuteTemplate(buffer, name, data); err != nil {
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
func (m *Msg) AddAlternativeHTMLTemplate(tpl *ht.Template, data any, opts ...PartOption) error {
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

// AddAlternativeNamedHTMLTemplate sets the body of the message from a template associated with a given
// html/template.Template pointer that has the given name.
//
// The content type will be set to "html/plain" automatically. This method executes the named template provided by text template
// with the given data and adds the result as an alternative version of the message body. If the template
// is nil or fails to execute, an error will be returned.
//
// Parameters:
//   - tpl: A pointer to the html/template.Template containing a named template.
//   - name: Name of the template to be used for the alternative body.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the alternative body part.
//
// Returns:
//   - An error if the template is nil or fails to execute, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2045
//   - https://datatracker.ietf.org/doc/html/rfc2046
func (m *Msg) AddAlternativeNamedHTMLTemplate(tpl *ht.Template, name string, data any, opts ...PartOption) error {
	if tpl == nil {
		return errors.New(errTplPointerNil)
	}
	buffer := bytes.NewBuffer(nil)
	if err := tpl.ExecuteTemplate(buffer, name, data); err != nil {
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
	name string, tpl *tt.Template, data any, opts ...FileOption,
) error {
	file, err := fileFromTextTemplate(name, tpl, data)
	if err != nil {
		return fmt.Errorf("failed to attach template: %w", err)
	}
	m.attachments = m.appendFile(m.attachments, file, opts...)
	return nil
}

// AttachNamedTextTemplate adds the output of a template associated with the text/template.Template pointer as a File attachment to the Msg.
//
// This method allows you to attach the rendered output of a text template as a file to the message.
// The named template associated with the provided text/template is executed with the provided data, and its output is attached as a file.
// If the template fails to execute, an error will be returned.
//
// Parameters:
//   - fileName: The name of the file to be attached.
//   - tpl: A pointer to the text/template.Template associated with the template to be executed for the attachment.
//   - tplName: A name of the template from tpl to be executed.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the attachment.
//
// Returns:
//   - An error if the template fails to execute or cannot be attached, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) AttachNamedTextTemplate(
	fileName string, tpl *tt.Template, tplName string, data any, opts ...FileOption,
) error {
	file, err := fileFromNamedTextTemplate(fileName, tpl, tplName, data)
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
	name string, tpl *ht.Template, data any, opts ...FileOption,
) error {
	file, err := fileFromHTMLTemplate(name, tpl, data)
	if err != nil {
		return fmt.Errorf("failed to attach template: %w", err)
	}
	m.attachments = m.appendFile(m.attachments, file, opts...)
	return nil
}

// AttachNamedHTMLTemplate adds the output of a template associated with the html/template.Template pointer as a File attachment to the Msg.
//
// This method allows you to attach the rendered output of a html template as a file to the message.
// The named template associated with the provided html/template is executed with the provided data, and its output is attached as a file.
// If the template fails to execute, an error will be returned.
//
// Parameters:
//   - fileName: The name of the file to be attached.
//   - tpl: A pointer to the html/template.Template associated with the template to be executed for the attachment.
//   - tplName: A name of the template from tpl to be executed.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the attachment.
//
// Returns:
//   - An error if the template fails to execute or cannot be attached, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) AttachNamedHTMLTemplate(
	fileName string, tpl *ht.Template, tplName string, data any, opts ...FileOption,
) error {
	file, err := fileFromNamedHTMLTemplate(fileName, tpl, tplName, data)
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
	name string, tpl *tt.Template, data any, opts ...FileOption,
) error {
	file, err := fileFromTextTemplate(name, tpl, data)
	if err != nil {
		return fmt.Errorf("failed to embed template: %w", err)
	}
	m.embeds = m.appendFile(m.embeds, file, opts...)
	return nil
}

// EmbedNamedTextTemplate adds the output of a template associated with the text/template.Template pointer as an embedded File to the Msg.
//
// This method embeds the rendered output of a text template into the email message. The named template associated with the provided text/template is
// executed with the provided data, and its output is embedded as a file. If the template fails to execute,
// an error will be returned.
//
// Parameters:
//   - fileName: The name of the embedded file.
//   - tpl: A pointer to the text/template.Template associated with the template to be executed for the embedded content.
//   - tplName: A name of the template from tpl to be executed.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the embedded file.
//
// Returns:
//   - An error if the template fails to execute or cannot be embedded, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) EmbedNamedTextTemplate(
	fileName string, tpl *tt.Template, tplName string, data any, opts ...FileOption,
) error {
	file, err := fileFromNamedTextTemplate(fileName, tpl, tplName, data)
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
	name string, tpl *ht.Template, data any, opts ...FileOption,
) error {
	file, err := fileFromHTMLTemplate(name, tpl, data)
	if err != nil {
		return fmt.Errorf("failed to embed template: %w", err)
	}
	m.embeds = m.appendFile(m.embeds, file, opts...)
	return nil
}

// EmbedNamedHTMLTemplate adds the output of a template associated with the html/template.Template pointer as an embedded File to the Msg.
//
// This method embeds the rendered output of a html template into the email message. The named template associated with the provided html/template is
// executed with the provided data, and its output is embedded as a file. If the template fails to execute,
// an error will be returned.
//
// Parameters:
//   - fileName: The name of the embedded file.
//   - tpl: A pointer to the html/template.Template associated with the template to be executed for the embedded content.
//   - tplName: A name of the template from tpl to be executed.
//   - data: The data to populate the template.
//   - opts: Optional parameters for customizing the embedded file.
//
// Returns:
//   - An error if the template fails to execute or cannot be embedded, otherwise nil.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func (m *Msg) EmbedNamedHTMLTemplate(
	fileName string, tpl *ht.Template, tplName string, data any, opts ...FileOption,
) error {
	file, err := fileFromNamedHTMLTemplate(fileName, tpl, tplName, data)
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
func fileFromTextTemplate(name string, tpl *tt.Template, data any) (*File, error) {
	if tpl == nil {
		return nil, errors.New(errTplPointerNil)
	}
	buffer := bytes.Buffer{}
	if err := tpl.Execute(&buffer, data); err != nil {
		return nil, fmt.Errorf(errTplExecuteFailed, err)
	}
	return fileFromReader(name, &buffer)
}

// fileFromNamedTextTemplate returns a File pointer from a named template associated with a given text/template.Template.
//
// This method executes the named template assosociated with the provided text template with the given data and creates a File structure
// representing the output. The rendered template content is stored in a buffer and then processed
// as a file attachment or embed.
//
// Parameters:
//   - fileName: The name of the file to be created from the template output.
//   - tpl: A pointer to the text/template.Template associated with the template to be executed.
//   - tplName: A name of the template from tpl to be executed.
//   - data: The data to populate the template.
//
// Returns:
//   - A pointer to the File structure representing the rendered template.
//   - An error if the template is nil or if it fails to execute.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func fileFromNamedTextTemplate(fileName string, tpl *tt.Template, tplName string, data any) (*File, error) {
	if tpl == nil {
		return nil, errors.New(errTplPointerNil)
	}
	buffer := bytes.Buffer{}
	if err := tpl.ExecuteTemplate(&buffer, tplName, data); err != nil {
		return nil, fmt.Errorf(errTplExecuteFailed, err)
	}
	return fileFromReader(fileName, &buffer)
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
func fileFromHTMLTemplate(name string, tpl *ht.Template, data any) (*File, error) {
	if tpl == nil {
		return nil, errors.New(errTplPointerNil)
	}
	buffer := bytes.Buffer{}
	if err := tpl.Execute(&buffer, data); err != nil {
		return nil, fmt.Errorf(errTplExecuteFailed, err)
	}
	return fileFromReader(name, &buffer)
}

// fileFromNamedHTMLTemplate returns a File pointer from a named template associated with a given html/template.Template.
//
// This method executes the named template assosociated with the provided html template with the given data and creates a File structure
// representing the output. The rendered template content is stored in a buffer and then processed
// as a file attachment or embed.
//
// Parameters:
//   - fileName: The name of the file to be created from the template output.
//   - tpl: A pointer to the html/template.Template associated with the template to be executed.
//   - tplName: A name of the template from tpl to be executed.
//   - data: The data to populate the template.
//
// Returns:
//   - A pointer to the File structure representing the rendered template.
//   - An error if the template is nil or if it fails to execute.
//
// References:
//   - https://datatracker.ietf.org/doc/html/rfc2183
func fileFromNamedHTMLTemplate(fileName string, tpl *ht.Template, tplName string, data any) (*File, error) {
	if tpl == nil {
		return nil, errors.New(errTplPointerNil)
	}
	buffer := bytes.Buffer{}
	if err := tpl.ExecuteTemplate(&buffer, tplName, data); err != nil {
		return nil, fmt.Errorf(errTplExecuteFailed, err)
	}
	return fileFromReader(fileName, &buffer)
}
