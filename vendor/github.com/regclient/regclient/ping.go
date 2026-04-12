package regclient

import (
	"context"

	"github.com/regclient/regclient/types/ping"
	"github.com/regclient/regclient/types/ref"
)

// Ping verifies access to a registry or equivalent.
func (rc *RegClient) Ping(ctx context.Context, r ref.Ref) (ping.Result, error) {
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return ping.Result{}, err
	}

	return schemeAPI.Ping(ctx, r)
}
