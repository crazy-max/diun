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
