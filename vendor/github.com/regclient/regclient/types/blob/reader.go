package blob

import (
	"fmt"
	"io"
	"strconv"
	"sync"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/internal/limitread"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/mediatype"
)

// Reader was previously an interface. A type alias is provided for upgrading.
type Reader = *BReader

// BReader is used to read blobs.
type BReader struct {
	BCommon
	readBytes int64
	reader    io.Reader
	origRdr   io.Reader
	digester  digest.Digester
	mu        sync.Mutex
}

// NewReader creates a new BReader.
func NewReader(opts ...Opts) *BReader {
	bc := blobConfig{}
	for _, opt := range opts {
		opt(&bc)
	}
	if bc.resp != nil {
		// extract headers and reader if other fields not passed
		if bc.header == nil {
			bc.header = bc.resp.Header
		}
		if bc.rdr == nil {
			bc.rdr = bc.resp.Body
		}
	}
	if bc.header != nil {
		// extract fields from header if descriptor not passed
		if bc.desc.MediaType == "" {
			bc.desc.MediaType = mediatype.Base(bc.header.Get("Content-Type"))
		}
		if bc.desc.Size == 0 {
			cl, _ := strconv.Atoi(bc.header.Get("Content-Length"))
			bc.desc.Size = int64(cl)
		}
		if bc.desc.Digest == "" {
			bc.desc.Digest, _ = digest.Parse(bc.header.Get("Docker-Content-Digest"))
		}
	}
	br := BReader{
		BCommon: BCommon{
			r:         bc.r,
			desc:      bc.desc,
			rawHeader: bc.header,
			resp:      bc.resp,
		},
		origRdr: bc.rdr,
	}
	if bc.rdr != nil {
		br.blobSet = true
		br.digester = br.desc.DigestAlgo().Digester()
		rdr := bc.rdr
		if br.desc.Size > 0 {
			rdr = &limitread.LimitRead{
				Reader: rdr,
				Limit:  br.desc.Size,
			}
		}
		br.reader = io.TeeReader(rdr, br.digester.Hash())
	}
	return &br
}

// Close attempts to close the reader and populates/validates the digest.
func (r *BReader) Close() error {
	if r == nil || r.origRdr == nil {
		return nil
	}
	// attempt to close if available in original reader
	bc, ok := r.origRdr.(io.Closer)
	if !ok {
		return nil
	}
	return bc.Close()
}

// RawBody returns the original body from the request.
func (r *BReader) RawBody() ([]byte, error) {
	return io.ReadAll(r)
}

// Read passes through the read operation while computing the digest and tracking the size.
func (r *BReader) Read(p []byte) (int, error) {
	if r == nil || r.reader == nil {
		return 0, fmt.Errorf("blob has no reader: %w", io.ErrUnexpectedEOF)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	size, err := r.reader.Read(p)
	r.readBytes = r.readBytes + int64(size)
	if err == io.EOF {
		// check/save size
		if r.desc.Size == 0 {
			r.desc.Size = r.readBytes
		} else if r.readBytes < r.desc.Size {
			err = fmt.Errorf("%w [expected %d, received %d]: %w", errs.ErrShortRead, r.desc.Size, r.readBytes, err)
		} else if r.readBytes > r.desc.Size {
			err = fmt.Errorf("%w [expected %d, received %d]: %w", errs.ErrSizeLimitExceeded, r.desc.Size, r.readBytes, err)
		}
		// check/save digest
		if r.desc.Digest.Validate() != nil {
			r.desc.Digest = r.digester.Digest()
		} else if r.desc.Digest != r.digester.Digest() {
			err = fmt.Errorf("%w [expected %s, calculated %s]: %w", errs.ErrDigestMismatch, r.desc.Digest.String(), r.digester.Digest().String(), err)
		}
	}
	return size, err
}

// Seek passes through the seek operation, reseting or invalidating the digest
func (r *BReader) Seek(offset int64, whence int) (int64, error) {
	if r == nil || r.origRdr == nil {
		return 0, fmt.Errorf("blob has no reader")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if offset == 0 && whence == io.SeekCurrent {
		return r.readBytes, nil
	}
	// cannot do an arbitrary seek and still digest without a lot more complication
	if offset != 0 || whence != io.SeekStart {
		return r.readBytes, fmt.Errorf("unable to seek to arbitrary position")
	}
	rdrSeek, ok := r.origRdr.(io.Seeker)
	if !ok {
		return r.readBytes, fmt.Errorf("Seek unsupported")
	}
	o, err := rdrSeek.Seek(offset, whence)
	if err != nil || o != 0 {
		return r.readBytes, err
	}
	// reset internal offset and digest calculation
	rdr := r.origRdr
	if r.desc.Size > 0 {
		rdr = &limitread.LimitRead{
			Reader: rdr,
			Limit:  r.desc.Size,
		}
	}
	r.digester = r.desc.DigestAlgo().Digester()
	r.reader = io.TeeReader(rdr, r.digester.Hash())
	r.readBytes = 0

	return 0, nil
}

// ToOCIConfig converts a BReader to a BOCIConfig.
func (r *BReader) ToOCIConfig() (*BOCIConfig, error) {
	if r == nil || !r.blobSet {
		return nil, fmt.Errorf("blob is not defined")
	}
	if r.readBytes != 0 {
		return nil, fmt.Errorf("unable to convert after read has been performed")
	}
	blobBody, err := io.ReadAll(r)
	errC := r.Close()
	if err != nil {
		return nil, fmt.Errorf("error reading image config for %s: %w", r.r.CommonName(), err)
	}
	if errC != nil {
		return nil, fmt.Errorf("error closing blob reader: %w", err)
	}
	return NewOCIConfig(
		WithDesc(r.desc),
		WithHeader(r.rawHeader),
		WithRawBody(blobBody),
		WithRef(r.r),
		WithResp(r.resp),
	), nil
}

// ToTarReader converts a BReader to a BTarReader
func (r *BReader) ToTarReader() (*BTarReader, error) {
	if r == nil || !r.blobSet {
		return nil, fmt.Errorf("blob is not defined")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.readBytes != 0 {
		return nil, fmt.Errorf("unable to convert after read has been performed")
	}
	return NewTarReader(
		WithDesc(r.desc),
		WithHeader(r.rawHeader),
		WithRef(r.r),
		WithResp(r.resp),
		WithReader(r.reader),
	), nil
}
