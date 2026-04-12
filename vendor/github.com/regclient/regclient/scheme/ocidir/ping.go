package ocidir

import (
	"context"
	"fmt"
	"os"

	"github.com/regclient/regclient/types/ping"
	"github.com/regclient/regclient/types/ref"
)

// Ping for an ocidir verifies access to read the path.
func (o *OCIDir) Ping(ctx context.Context, r ref.Ref) (ping.Result, error) {
	ret := ping.Result{}
	fd, err := os.Open(r.Path)
	if err != nil {
		return ret, err
	}
	defer fd.Close()
	fi, err := fd.Stat()
	if err != nil {
		return ret, err
	}
	ret.Stat = fi
	if !fi.IsDir() {
		return ret, fmt.Errorf("failed to access %s: not a directory", r.Path)
	}
	return ret, nil
}
