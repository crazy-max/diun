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

//go:build !go1.18

package version

import (
	"fmt"
	"runtime"
)

type Info struct {
	GoVer      string `json:"goVersion"`  // go version
	GoCompiler string `json:"goCompiler"` // go compiler
	Platform   string `json:"platform"`   // os/arch
	VCSCommit  string `json:"vcsCommit"`  // commit sha
	VCSDate    string `json:"vcsDate"`    // commit date in RFC3339 format
	VCSRef     string `json:"vcsRef"`     // commit sha + dirty if state is not clean
	VCSState   string `json:"vcsState"`   // clean or dirty
	VCSTag     string `json:"vcsTag"`     // tag
}

func GetInfo() Info {
	i := Info{
		GoVer:     unknown,
		Platform:  unknown,
		VCSCommit: unknown,
		VCSDate:   unknown,
		VCSRef:    unknown,
		VCSState:  unknown,
		VCSTag:    "",
	}

	i.GoVer = runtime.Version()
	i.GoCompiler = runtime.Compiler
	i.Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	return i
}
