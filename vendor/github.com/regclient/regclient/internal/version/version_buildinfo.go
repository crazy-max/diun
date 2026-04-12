//go:build go1.18

package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

var vcsTag = ""

type Info struct {
	GoVer      string           `json:"goVersion"`       // go version
	GoCompiler string           `json:"goCompiler"`      // go compiler
	Platform   string           `json:"platform"`        // os/arch
	VCSCommit  string           `json:"vcsCommit"`       // commit sha
	VCSDate    string           `json:"vcsDate"`         // commit date in RFC3339 format
	VCSRef     string           `json:"vcsRef"`          // commit sha + dirty if state is not clean
	VCSState   string           `json:"vcsState"`        // clean or dirty
	VCSTag     string           `json:"vcsTag"`          // tag is not available from Go
	Debug      *debug.BuildInfo `json:"debug,omitempty"` // build info debugging data
}

func GetInfo() Info {
	i := Info{
		GoVer:     unknown,
		Platform:  unknown,
		VCSCommit: unknown,
		VCSDate:   unknown,
		VCSRef:    unknown,
		VCSState:  unknown,
		VCSTag:    vcsTag,
	}

	i.GoVer = runtime.Version()
	i.GoCompiler = runtime.Compiler
	i.Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
		i.Debug = bi
		if i.VCSTag == "" {
			i.VCSTag = bi.Main.Version
		}
		date := biSetting(bi, biVCSDate)
		if t, err := time.Parse(time.RFC3339, date); err == nil {
			i.VCSDate = t.UTC().Format(time.RFC3339)
		}
		i.VCSCommit = biSetting(bi, biVCSCommit)
		i.VCSRef = i.VCSCommit
		modified := biSetting(bi, biVCSModified)
		if modified == "true" {
			i.VCSState = stateDirty
			i.VCSRef += "-" + stateDirty
		} else if modified == "false" {
			i.VCSState = stateClean
		}
	}

	return i
}

func biSetting(bi *debug.BuildInfo, key string) string {
	if bi == nil {
		return unknown
	}
	for _, setting := range bi.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return unknown
}
