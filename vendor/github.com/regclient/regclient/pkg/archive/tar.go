package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"time"
)

// TarOpts configures options for Create/Extract tar
type TarOpts func(*tarOpts)

// TODO: add support for compressed files with bzip
type tarOpts struct {
	// allowRelative bool // allow relative paths outside of target folder
	compress string
}

// TarCompressGzip option to use gzip compression on tar files
func TarCompressGzip(to *tarOpts) {
	to.compress = "gzip"
}

// TarUncompressed option to tar (noop)
func TarUncompressed(to *tarOpts) {
}

// TODO: add option for full path or to adjust the relative path

// Tar creation
func Tar(ctx context.Context, path string, w io.Writer, opts ...TarOpts) error {
	to := tarOpts{}
	for _, opt := range opts {
		opt(&to)
	}

	twOut := w
	if to.compress == "gzip" {
		gw := gzip.NewWriter(w)
		defer gw.Close()
		twOut = gw
	}

	tw := tar.NewWriter(twOut)
	defer tw.Close()

	// walk the path performing a recursive tar
	err := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		// return any errors filepath encounters accessing the file
		if err != nil {
			return err
		}

		// TODO: handle symlinks, security attributes, hard links
		// TODO: add options for file owner and timestamps
		// TODO: add options to override time, or disable access/change stamps

		// adjust for relative path
		relPath, err := filepath.Rel(path, file)
		if err != nil || relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, relPath)
		if err != nil {
			return err
		}

		header.Format = tar.FormatPAX
		header.Name = filepath.ToSlash(relPath)
		header.AccessTime = time.Time{}
		header.ChangeTime = time.Time{}
		header.ModTime = header.ModTime.Truncate(time.Second)

		if err = tw.WriteHeader(header); err != nil {
			return err
		}

		// open file and copy contents into tar writer
		if header.Typeflag == tar.TypeReg && header.Size > 0 {
			//#nosec G304 filename is limited to provided path directory
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err = io.Copy(tw, f); err != nil {
				return err
			}
			err = f.Close()
			if err != nil {
				return fmt.Errorf("failed to close file: %w", err)
			}
		}
		return nil
	})
	return err
}

// Extract Tar
func Extract(ctx context.Context, path string, r io.Reader, opts ...TarOpts) error {
	to := tarOpts{}
	for _, opt := range opts {
		opt(&to)
	}

	// verify path exists
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("extract path must be a directory: \"%s\"", path)
	}

	// decompress
	rd, err := Decompress(r)
	if err != nil {
		return err
	}

	rt := tar.NewReader(rd)
	for {
		hdr, err := rt.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// join a cleaned version of the filename with the path
		fn := filepath.Join(path, filepath.Clean("/"+hdr.Name))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if hdr.Mode < 0 || hdr.Mode > math.MaxUint32 {
				return fmt.Errorf("integer conversion overflow/underflow (file mode = %d)", hdr.Mode)
			}
			err = os.MkdirAll(fn, fs.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
		case tar.TypeReg:
			// TODO: configure file mode, creation timestamp, etc
			//#nosec G304 filename is limited to provided path directory
			fh, err := os.Create(fn)
			if err != nil {
				return err
			}
			n, err := io.CopyN(fh, rt, hdr.Size)
			errC := fh.Close()
			if err != nil {
				return err
			}
			if errC != nil {
				return fmt.Errorf("failed to close file: %w", errC)
			}
			if n != hdr.Size {
				return fmt.Errorf("size mismatch extracting \"%s\", expected %d, extracted %d", hdr.Name, hdr.Size, n)
			}
			// TODO: handle other tar types (symlinks, etc)
		}
	}

	return nil
}
