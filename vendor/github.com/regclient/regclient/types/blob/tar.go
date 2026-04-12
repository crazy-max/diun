package blob

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/internal/limitread"
	"github.com/regclient/regclient/pkg/archive"
	"github.com/regclient/regclient/types/errs"
)

// TarReader was previously an interface. A type alias is provided for upgrading.
type TarReader = *BTarReader

// BTarReader is used to read individual files from an image layer.
type BTarReader struct {
	BCommon
	origRdr  io.Reader
	reader   io.Reader
	digester digest.Digester
	tr       *tar.Reader
}

// NewTarReader creates a BTarReader.
// Typically a BTarReader will be created using BReader.ToTarReader().
func NewTarReader(opts ...Opts) *BTarReader {
	bc := blobConfig{}
	for _, opt := range opts {
		opt(&bc)
	}
	tr := BTarReader{
		BCommon: BCommon{
			desc:      bc.desc,
			r:         bc.r,
			rawHeader: bc.header,
			resp:      bc.resp,
		},
		origRdr: bc.rdr,
	}
	if bc.rdr != nil {
		tr.blobSet = true
		tr.digester = tr.desc.DigestAlgo().Digester()
		rdr := bc.rdr
		if tr.desc.Size > 0 {
			rdr = &limitread.LimitRead{
				Reader: rdr,
				Limit:  tr.desc.Size,
			}
		}
		tr.reader = io.TeeReader(rdr, tr.digester.Hash())
	}
	return &tr
}

// Close attempts to close the reader and populates/validates the digest.
func (tr *BTarReader) Close() error {
	// attempt to close if available in original reader
	if trc, ok := tr.origRdr.(io.Closer); ok && trc != nil {
		return trc.Close()
	}
	return nil
}

// GetTarReader returns the tar.Reader for the blob.
func (tr *BTarReader) GetTarReader() (*tar.Reader, error) {
	if tr.reader == nil {
		return nil, fmt.Errorf("blob has no reader defined")
	}
	if tr.tr == nil {
		dr, err := archive.Decompress(tr.reader)
		if err != nil {
			return nil, err
		}
		tr.tr = tar.NewReader(dr)
	}
	return tr.tr, nil
}

// RawBody returns the original body from the request.
func (tr *BTarReader) RawBody() ([]byte, error) {
	if !tr.blobSet {
		return []byte{}, fmt.Errorf("Blob is not defined")
	}
	if tr.tr != nil {
		return []byte{}, fmt.Errorf("RawBody cannot be returned after TarReader returned")
	}
	b, err := io.ReadAll(tr.reader)
	if err != nil {
		return b, err
	}
	if tr.digester != nil {
		dig := tr.digester.Digest()
		tr.digester = nil
		if tr.desc.Digest.String() != "" && dig != tr.desc.Digest {
			return b, fmt.Errorf("%w, expected %s, received %s", errs.ErrDigestMismatch, tr.desc.Digest.String(), dig.String())
		}
		tr.desc.Digest = dig
	}
	err = tr.Close()
	return b, err
}

// ReadFile parses the tar to find a file.
func (tr *BTarReader) ReadFile(filename string) (*tar.Header, io.Reader, error) {
	if strings.HasPrefix(filename, ".wh.") {
		return nil, nil, fmt.Errorf(".wh. prefix is reserved for whiteout files")
	}
	// normalize filenames,
	filename = filepath.Clean(filename)
	if filename[0] == '/' {
		filename = filename[1:]
	}
	// get reader
	rdr, err := tr.GetTarReader()
	if err != nil {
		return nil, nil, err
	}
	// loop through files until whiteout or target file is found
	whiteout := false
	for {
		th, err := rdr.Next()
		if err != nil {
			// break on eof, everything else is an error
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, nil, err
		}
		thFile := filepath.Clean(th.Name)
		if thFile[0] == '/' {
			thFile = thFile[1:]
		}
		// found the target file
		if thFile == filename {
			return th, rdr, nil
		}
		// check/track whiteout file
		name := filepath.Base(th.Name)
		if !whiteout && strings.HasPrefix(name, ".wh.") && tarCmpWhiteout(th.Name, filename) {
			// continue searching after finding a whiteout file
			// a new file may be created in the same layer
			whiteout = true
		}
	}
	// EOF encountered
	if whiteout {
		return nil, nil, errs.ErrFileDeleted
	}
	if tr.digester != nil {
		_, _ = io.Copy(io.Discard, tr.reader) // process/digest any trailing bytes from reader
		dig := tr.digester.Digest()
		tr.digester = nil
		if tr.desc.Digest.String() != "" && dig != tr.desc.Digest {
			return nil, nil, fmt.Errorf("%w, expected %s, received %s", errs.ErrDigestMismatch, tr.desc.Digest.String(), dig.String())
		}
		tr.desc.Digest = dig
	}
	return nil, nil, errs.ErrFileNotFound
}

func tarCmpWhiteout(whFile, tgtFile string) bool {
	whSplit := strings.Split(whFile, "/")
	tgtSplit := strings.Split(tgtFile, "/")
	// the -1 handles the opaque whiteout
	if len(whSplit)-1 > len(tgtSplit) {
		return false
	}
	// verify the path matches up to the whiteout
	for i := range whSplit[:len(whSplit)-1] {
		if whSplit[i] != tgtSplit[i] {
			return false
		}
	}
	i := len(whSplit) - 1
	// opaque whiteout of entire directory
	if whSplit[i] == ".wh..wh..opq" {
		return true
	}
	// compare whiteout name to next path entry
	if i > len(tgtSplit)-1 {
		return false
	}
	whName := strings.TrimPrefix(whSplit[i], ".wh.")
	return whName == tgtSplit[i]
}
