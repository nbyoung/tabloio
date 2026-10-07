package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// guardOutput refuses a path with a .tableaux or a .git component: the
// output of a read command never lands in the project's state. It reads the
// components as written, before any cleaning, so that a path which climbs
// out of such a directory again is refused too.
func guardOutput(path string) error {
	for _, c := range strings.Split(filepath.ToSlash(path), "/") {
		if c == ".tableaux" || c == ".git" {
			return fmt.Errorf("--output does not write inside %s", c)
		}
	}
	return nil
}

// resolveOutput returns the path the output lands at: out as given, or
// joined to dir, the value of -C, when out is relative.
func resolveOutput(dir, out string) string {
	if dir == "" || filepath.IsAbs(out) {
		return out
	}
	return filepath.Join(dir, out)
}

// linkBase returns the path from the directory that holds the output file
// out to the project directory dir, with forward slashes, and "" when they
// are the same directory. A relative dir is the working directory's own, ""
// included. When no relative path exists, as between two volumes, it
// returns the absolute path of dir.
func linkBase(dir, out string) string {
	if dir == "" {
		dir = "."
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return filepath.ToSlash(dir)
	}
	absOut, err := filepath.Abs(out)
	if err != nil {
		return filepath.ToSlash(absDir)
	}
	rel, err := filepath.Rel(filepath.Dir(absOut), absDir)
	if err != nil {
		return filepath.ToSlash(absDir)
	}
	if rel == "." {
		return ""
	}
	return filepath.ToSlash(rel)
}

// writeFile writes data to path through a temporary file in the same
// directory and a rename, with mode 0644, so that a failure leaves an
// existing file as it stood and no temporary file behind.
func writeFile(path string, data []byte) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tabloio-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		return err
	}
	if err = tmp.Chmod(0o644); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
