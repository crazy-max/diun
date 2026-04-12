package archive

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// CompressType identifies the detected compression type
type CompressType int

const (
	CompressNone  CompressType = iota // uncompressed or unable to detect compression
	CompressBzip2                     // bzip2
	CompressGzip                      // gzip
	CompressXz                        // xz
	CompressZstd                      // zstd
)

// compressHeaders are used to detect the compression type
var compressHeaders = map[CompressType][]byte{
	CompressBzip2: []byte("\x42\x5A\x68"),
	CompressGzip:  []byte("\x1F\x8B\x08"),
	CompressXz:    []byte("\xFD\x37\x7A\x58\x5A\x00"),
	CompressZstd:  []byte("\x28\xB5\x2F\xFD"),
}

func Compress(r io.Reader, oComp CompressType) (io.ReadCloser, error) {
	switch oComp {
	// note, bzip2 compression is not supported
	case CompressGzip:
		return writeToRead(r, newGzipWriter)
	case CompressXz:
		return writeToRead(r, xz.NewWriter)
	case CompressZstd:
		return writeToRead(r, newZstdWriter)
	case CompressNone:
		return io.NopCloser(r), nil
	default:
		return nil, ErrUnknownType
	}
}

// newGzipWriter generates a writer and an always nil error.
func newGzipWriter(w io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

// newZstdWriter generates a writer with the default options.
func newZstdWriter(w io.Writer) (io.WriteCloser, error) {
	return zstd.NewWriter(w)
}

// writeToRead uses a pipe + goroutine + copy to switch from a writer to a reader.
func writeToRead[wc io.WriteCloser](src io.Reader, newWriterFn func(io.Writer) (wc, error)) (io.ReadCloser, error) {
	pr, pw := io.Pipe()
	go func() {
		// buffer output to avoid lots of small reads
		bw := bufio.NewWriterSize(pw, 2<<16)
		dest, err := newWriterFn(bw)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(dest, src); err != nil {
			_ = pw.CloseWithError(err)
		}
		if err := dest.Close(); err != nil {
			_ = pw.CloseWithError(err)
		}
		if err := bw.Flush(); err != nil {
			_ = pw.CloseWithError(err)
		}
		_ = pw.Close()
	}()
	return pr, nil
}

// Decompress extracts gzip and bzip streams
func Decompress(r io.Reader) (io.Reader, error) {
	// create bufio to peak on first few bytes
	br := bufio.NewReader(r)
	head, err := br.Peek(10)
	if err != nil && !errors.Is(err, io.EOF) {
		return br, fmt.Errorf("failed to detect compression: %w", err)
	}

	// compare peaked data against known compression types
	switch DetectCompression(head) {
	case CompressBzip2:
		return bzip2.NewReader(br), nil
	case CompressGzip:
		return gzip.NewReader(br)
	case CompressXz:
		return xz.NewReader(br)
	case CompressZstd:
		return zstd.NewReader(br)
	default:
		return br, nil
	}
}

// DetectCompression identifies the compression type based on the first few bytes
func DetectCompression(head []byte) CompressType {
	for c, b := range compressHeaders {
		if bytes.HasPrefix(head, b) {
			return c
		}
	}
	return CompressNone
}

func (ct CompressType) String() string {
	mt, err := ct.MarshalText()
	if err != nil {
		return "unknown"
	}
	return string(mt)
}

func (ct CompressType) MarshalText() ([]byte, error) {
	switch ct {
	case CompressNone:
		return []byte("none"), nil
	case CompressBzip2:
		return []byte("bzip2"), nil
	case CompressGzip:
		return []byte("gzip"), nil
	case CompressXz:
		return []byte("xz"), nil
	case CompressZstd:
		return []byte("zstd"), nil
	}
	return nil, fmt.Errorf("unknown compression type")
}

func (ct *CompressType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "none":
		*ct = CompressNone
	case "bzip2":
		*ct = CompressBzip2
	case "gzip":
		*ct = CompressGzip
	case "xz":
		*ct = CompressXz
	case "zstd":
		*ct = CompressZstd
	default:
		return fmt.Errorf("unknown compression type %s", string(text))
	}
	return nil
}
