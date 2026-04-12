package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// credHelper wraps a command that manages user credentials.
type credHelper struct {
	prog string
	env  map[string]string
}

func newCredHelper(prog string, env map[string]string) *credHelper {
	return &credHelper{prog: prog, env: env}
}

func (ch *credHelper) run(arg string, input io.Reader) ([]byte, error) {
	//#nosec G204 only untrusted arg is a hostname which the executed command should not trust
	cmd := exec.Command(ch.prog, arg)
	cmd.Env = os.Environ()
	if ch.env != nil {
		for k, v := range ch.env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	cmd.Stderr = os.Stderr
	cmd.Stdin = input
	return cmd.Output()
}

type credStore struct {
	ServerURL string `json:"ServerURL"`
	Username  string `json:"Username"`
	Secret    string `json:"Secret"` //#nosec G117 exported struct intentionally holds secrets
}

// get requests a credential from the helper for a given host.
func (ch *credHelper) get(host *Host) error {
	hostname := host.Hostname
	if host.CredHost != "" {
		hostname = host.CredHost
	}
	hostIn := strings.NewReader(hostname)
	credOut := credStore{
		Username: host.User,
		Secret:   host.Pass,
	}
	outB, err := ch.run("get", hostIn)
	if err != nil {
		outS := strings.TrimSpace(string(outB))
		return fmt.Errorf("error getting credentials, output: %s, error: %w", outS, err)
	}
	err = json.NewDecoder(bytes.NewReader(outB)).Decode(&credOut)
	if err != nil {
		return fmt.Errorf("error reading credentials: %w", err)
	}
	if credOut.Username == tokenUser {
		host.User = ""
		host.Pass = ""
		host.Token = credOut.Secret
	} else {
		host.User = credOut.Username
		host.Pass = credOut.Secret
		host.Token = ""
	}
	return nil
}

// list returns a list of hosts supported by the credential helper.
func (ch *credHelper) list() ([]Host, error) {
	credList := map[string]string{}
	outB, err := ch.run("list", bytes.NewReader([]byte{}))
	if err != nil {
		outS := strings.TrimSpace(string(outB))
		return nil, fmt.Errorf("error getting credential list, output: %s, error: %w", outS, err)
	}
	err = json.NewDecoder(bytes.NewReader(outB)).Decode(&credList)
	if err != nil {
		return nil, fmt.Errorf("error reading credential list: %w", err)
	}
	hostList := []Host{}
	for host, user := range credList {
		if !HostValidate(host) {
			continue
		}
		h := HostNewName(host)
		h.User = user
		h.CredHelper = ch.prog
		hostList = append(hostList, *h)
	}
	return hostList, nil
}

// TODO: store method not implemented
