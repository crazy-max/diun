package secret

import "os"

// GetSecret retrieves secret's value from plaintext or filename if defined.
func GetSecret(plaintext, filename string) (string, error) {
	if plaintext != "" {
		return plaintext, nil
	}
	if filename != "" {
		b, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", nil
}
