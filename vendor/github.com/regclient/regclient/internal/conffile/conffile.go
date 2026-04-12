// Package conffile wraps the read and write of configuration files
package conffile

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
)

type File struct {
	perms    int
	fullname string
}

type Opt func(*File)

// New returns a new File.
// The last successful option determines the filename.
func New(opts ...Opt) *File {
	f := File{perms: 0o600}
	for _, fn := range opts {
		fn(&f)
	}
	if f.fullname == "" {
		return nil
	}
	return &f
}

// WithAppDir determines the filename from the XDG or Windows specification.
// By default, this is based in $HOME/.config on Linux and %APPDATA% on Windows.
// If the file does not exist, this will set the filename only if "force" is true.
func WithAppDir(unixDir, winDir, name string, force bool) Opt {
	var dir string
	if winDir == "" {
		dir = unixDir
	} else {
		dir = osString(unixDir, winDir)
	}
	return func(f *File) {
		fullname := filepath.Join(appDir(), dir, name)
		if force || exists(fullname) {
			f.fullname = fullname
		}
	}
}

// WithDirName determines the filename from a subdirectory in the user's HOME.
//
// Deprecated: Replace with [WithHomeDir]
//
//go:fix inline
func WithDirName(dir, name string) Opt {
	return WithHomeDir(dir, name, true)
}

// WithEnvFile sets the fullname to the environment value if defined.
func WithEnvFile(envVar string) Opt {
	return func(f *File) {
		val := os.Getenv(envVar)
		if val != "" {
			f.fullname = val
		}
	}
}

// WithEnvDir sets the fullname to the environment value + filename if the environment variable is defined.
func WithEnvDir(envVar, name string) Opt {
	return func(f *File) {
		val := os.Getenv(envVar)
		if val != "" {
			f.fullname = filepath.Join(val, name)
		}
	}
}

// WithFullname specifies the filename.
// This will always set the filename even if the file does not exist.
func WithFullname(fullname string) Opt {
	return func(f *File) {
		f.fullname = fullname
	}
}

// WithHomeDir determines the filename from a subdirectory in the user's HOME
// e.g. dir=".app", name="config.json", sets the fullname to "$HOME/.app/config.json".
// If the file does not exist, this will set the filename only if "force" is true.
func WithHomeDir(dir, name string, force bool) Opt {
	return func(f *File) {
		filename := filepath.Join(homeDir(), dir, name)
		if force || exists(filename) {
			f.fullname = filename
		}
	}
}

// WithPerms specifies the permissions to create a file with (default 0600).
func WithPerms(perms int) Opt {
	return func(f *File) {
		f.perms = perms
	}
}

func (f *File) Name() string {
	return f.fullname
}

func (f *File) Open() (io.ReadCloser, error) {
	return os.Open(f.fullname)
}

func (f *File) Write(rdr io.Reader) error {
	// create temp file/open
	dir := filepath.Dir(f.fullname)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(f.fullname))
	if err != nil {
		return err
	}
	tmpStat, err := tmp.Stat()
	if err != nil {
		return err
	}
	tmpName := tmpStat.Name()
	tmpFullname := filepath.Join(dir, tmpName)
	defer os.Remove(tmpFullname)

	// copy from rdr to temp file
	_, err = io.Copy(tmp, rdr)
	errC := tmp.Close()
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	if errC != nil {
		return fmt.Errorf("failed to close config: %w", errC)
	}

	// adjust file ownership/permissions
	mode := os.FileMode(0o600)
	uid := os.Getuid()
	gid := os.Getgid()
	// adjust defaults based on existing file if available
	stat, err := os.Stat(f.fullname)
	if err == nil {
		// adjust mode to existing file
		if stat.Mode().IsRegular() {
			mode = stat.Mode()
		}
		uid, gid, _ = getFileOwner(stat)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	// update mode and owner of temp file
	//#nosec G703 tempfile location is user controlled
	if err := os.Chmod(tmpFullname, mode); err != nil {
		return err
	}
	if uid > 0 && gid > 0 {
		//#nosec G703 tempfile location is user controlled
		_ = os.Chown(tmpFullname, uid, gid)
	}
	// move temp file to target filename
	//#nosec G703 tempfile location is user controlled
	return os.Rename(tmpFullname, f.fullname)
}

func exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func homeDir() string {
	home := os.Getenv(homeEnv)
	if home == "" {
		u, err := user.Current()
		if err == nil {
			home = u.HomeDir
		}
	}
	return home
}
