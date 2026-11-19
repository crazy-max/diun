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

package reproducible

import (
	"errors"
	"os"
	"strconv"
	"time"
)

const EpocEnv = "SOURCE_DATE_EPOC"

var errInvalidEpoc = errors.New("invalid epoc var")

// TimeNow returns the current time or SOURCE_DATE_EPOC if that is set.
func TimeNow() time.Time {
	now, err := TimeEpocEnv()
	if err == nil {
		return now
	}
	return time.Now().UTC()
}

// TimeEpocEnv returns the time parsed by SOURCE_DATE_EPOC.
// This should be used to override any timestamps that should be reproducible.
func TimeEpocEnv() (time.Time, error) {
	sec := os.Getenv(EpocEnv)
	if sec == "" {
		return time.Time{}, errInvalidEpoc
	}
	secI, err := strconv.ParseInt(sec, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(secI, 0).UTC(), nil
}
