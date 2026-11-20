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

//go:build windows

package conffile

import (
	"io/fs"
	"os"
	"path/filepath"
)

const (
	appDirEnv = "APPDATA"
	homeEnv   = "USERPROFILE"
)

func appDir() string {
	appDir := os.Getenv(appDirEnv)
	if appDir == "" {
		home := homeDir()
		appDir = filepath.Join(home, "AppData")
	}
	return appDir
}

func getFileOwner(_ fs.FileInfo) (int, int, error) {
	return 0, 0, nil
}

func osString(_, win string) string {
	return win
}
