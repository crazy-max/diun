package gotgbot

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
)

// InputFile (https://core.telegram.org/bots/api#inputfile)
//
// This object represents the contents of a file to be uploaded.
// Must be posted using multipart/form-data in the usual way that files are uploaded via the browser.
type InputFile interface {
	InputFileOrString
	justFiles()
}

// InputFileOrString (https://core.telegram.org/bots/api#inputfile)
//
// This object represents the contents of a file to be uploaded, or a publicly accessible URL to be reused.
// Files must be posted using multipart/form-data in the usual way that files are uploaded via the browser.
type InputFileOrString interface {
	Attach(name string, w *multipart.Writer) error
	getValue() string
}

var (
	_ InputFileOrString = &FileReader{}
	_ InputFile         = &FileReader{}
)

type FileReader struct {
	Name string
	Data io.Reader

	value string
}

func (f *FileReader) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.getValue())
}

func (f *FileReader) justFiles() {}

// Attach will handle any requirements to write any files to the multipartwriter.
// If FileReader.Data is nil, nothing happens.
// if FileReader.Data is an io.Seeker, then we seek the start of the file to ensure that the entire thing is sent.
// This also ensures that the request can be seamlessly retried.
// A Seeker interface can be easily obtained by using an *os.File, *bytes.Reader, or *strings.Reader.
func (f *FileReader) Attach(key string, w *multipart.Writer) error {
	if f.Data == nil {
		// if no data, this must be a string; nothing to "attach".
		return nil
	}

	fileName := f.Name
	if fileName == "" {
		fileName = key
	}

	part, err := w.CreateFormFile(key, fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file for field %s and fileName %s: %w", key, fileName, err)
	}

	if seeker, ok := f.Data.(io.Seeker); ok {
		// If this is a seeker, then we reset to the start of the file, to ensure retries work as expected.
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek file for field %s and fileName %s: %w", key, fileName, err)
		}
	}

	_, err = io.Copy(part, f.Data)
	if err != nil {
		return fmt.Errorf("failed to copy file contents of field %s to form: %w", key, err)
	}
	f.value = "attach://" + key
	return nil
}

// getValue returns the file attach reference for the relevant multipart form.
// Make sure to only call getValue after having called Attach(), to ensure any files have been included.
func (f *FileReader) getValue() string {
	return f.value
}

// InputFileByURL is used to send a file on the internet via a publicly accessible HTTP URL.
func InputFileByURL(url string) InputFileOrString {
	return &FileReader{value: url}
}

// InputFileByID is used to send a file that is already present on telegram's servers, using its telegram file_id.
func InputFileByID(fileID string) InputFileOrString {
	return &FileReader{value: fileID}
}

// InputFileByReader is used to send a file by a reader interface; such as a filehandle from os.Open(), or from a byte
// buffer.
//
// For example:
//
//	f, err := os.Open("some_file.go")
//	if err != nil {
//		return fmt.Errorf("failed to open file: %w", err)
//	}
//
//	m, err := b.SendDocument(<chat_id>, gotgbot.InputFileByReader("source.go", f), nil)
//
// Or
//
//	m, err := b.SendDocument(<chat_id>, gotgbot.InputFileByReader("file.txt", strings.NewReader("Some file contents")), nil)
func InputFileByReader(name string, r io.Reader) InputFile {
	return &FileReader{Name: name, Data: r}
}
