// Package limitread provides a reader that will error if the limit is ever exceeded
package limitread

import (
	"fmt"
	"io"

	"github.com/regclient/regclient/types/errs"
)

type LimitRead struct {
	Reader io.Reader
	Limit  int64
}

func (lr *LimitRead) Read(p []byte) (int, error) {
	if lr.Limit < 0 {
		return 0, fmt.Errorf("read limit exceeded%.0w", errs.ErrSizeLimitExceeded)
	}
	if int64(len(p)) > lr.Limit+1 {
		p = p[0 : lr.Limit+1]
	}
	n, err := lr.Reader.Read(p)
	lr.Limit -= int64(n)
	if lr.Limit < 0 {
		return n, fmt.Errorf("read limit exceeded%.0w", errs.ErrSizeLimitExceeded)
	}
	return n, err
}
