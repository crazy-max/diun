// Package timejson extends time methods with marshal/unmarshal for json
package timejson

import (
	"encoding/json"
	"errors"
	"time"
)

var errInvalid = errors.New("invalid duration")

// Duration is an alias to time.Duration
// Implementation taken from https://stackoverflow.com/questions/48050945/how-to-unmarshal-json-into-durations
type Duration time.Duration

// MarshalJSON converts a duration to json
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON converts json to a duration
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		timeDur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(timeDur)
		return nil
	default:
		return errInvalid
	}
}
