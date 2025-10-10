package utl

import (
	"os"
	"regexp"
	"time"
)

// MatchString reports whether a string s
// contains any match of a regular expression.
func MatchString(exp string, s string) bool {
	re, err := regexp.Compile(exp)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// IsIncluded checks if s string is included in includes
// If includes is empty, assume true
func IsIncluded(s string, includes []string) bool {
	if len(includes) == 0 {
		return true
	}
	for _, include := range includes {
		if MatchString(include, s) {
			return true
		}
	}
	return false
}

// IsExcluded checks if s string is excluded in excludes
// If excludes is empty, assume false
func IsExcluded(s string, excludes []string) bool {
	if len(excludes) == 0 {
		return false
	}
	for _, exclude := range excludes {
		if MatchString(exclude, s) {
			return true
		}
	}
	return false
}

// GetEnv retrieves the value of the environment variable named by the key
// or fallback if not found
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetSecret retrieves secret's value from plaintext or filename if defined
func GetSecret(plaintext, filename string) (string, error) {
	if plaintext != "" {
		return plaintext, nil
	} else if filename != "" {
		b, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", nil
}

// GetTemplate retrieves template's value from plaintext or filename if defined
func GetTemplate(plaintext, filename string) (string, error) {
	if plaintext != "" {
		return plaintext, nil
	} else if filename != "" {
		b, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", nil
}

// NewFalse returns a false bool pointer
func NewFalse() *bool {
	b := false
	return &b
}

// NewTrue returns a true bool pointer
func NewTrue() *bool {
	b := true
	return &b
}

// NewDuration returns a duration pointer
func NewDuration(duration time.Duration) *time.Duration {
	return &duration
}

// Contains checks if a slice contains a string
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
