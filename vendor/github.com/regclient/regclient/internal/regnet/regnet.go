// Copyright the regclient contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package regnet contains networking helper functions for interacting with registries.
package regnet

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/regclient/regclient/types/errs"
)

func AllowRedirect(src, dest url.URL) error {
	if src.Scheme == "https" && dest.Scheme != "https" {
		return fmt.Errorf("redirect from an https to non-https server is not allowed (%s)%.0w", dest.String(), errs.ErrHTTPRedirectRefused)
	}
	if !IsLocal(src.Host) && IsLocal(dest.Host) {
		return fmt.Errorf("redirect to a local domain is not allowed (%s)%.0w", dest.String(), errs.ErrHTTPRedirectRefused)
	}
	return nil
}

func IsLocal(hostPort string) bool {
	// skip check on any requests going through a proxy
	if u, err := http.ProxyFromEnvironment(&http.Request{URL: &url.URL{Scheme: "https", Host: hostPort}}); err == nil && u != nil {
		return false
	}
	// strip trailing port
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		host = hostPort
	}
	// parse IP
	ip := net.ParseIP(host)
	if ip != nil {
		return isIPLocal(ip)
	}
	// else resolve the hostname and then check each IP
	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}
	for _, ip := range ips {
		if ip != nil && isIPLocal(ip) {
			return true
		}
	}
	return false
}

func isIPLocal(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
