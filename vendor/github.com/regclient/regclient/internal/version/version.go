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

// Package version returns details on the Go and Git repo used in the build
package version

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

const (
	stateClean    = "clean"
	stateDirty    = "dirty"
	unknown       = "unknown"
	biVCSDate     = "vcs.time"
	biVCSCommit   = "vcs.revision"
	biVCSModified = "vcs.modified"
)

func (i Info) MarshalPretty() ([]byte, error) {
	buf := &bytes.Buffer{}
	tw := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	fmt.Fprintf(tw, "VCSTag:\t%s\n", i.VCSTag)
	fmt.Fprintf(tw, "VCSRef:\t%s\n", i.VCSRef)
	fmt.Fprintf(tw, "VCSCommit:\t%s\n", i.VCSCommit)
	fmt.Fprintf(tw, "VCSState:\t%s\n", i.VCSState)
	fmt.Fprintf(tw, "VCSDate:\t%s\n", i.VCSDate)
	fmt.Fprintf(tw, "Platform:\t%s\n", i.Platform)
	fmt.Fprintf(tw, "GoVer:\t%s\n", i.GoVer)
	fmt.Fprintf(tw, "GoCompiler:\t%s\n", i.GoCompiler)
	err := tw.Flush()
	return buf.Bytes(), err
}
