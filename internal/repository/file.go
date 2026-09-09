package repository

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const DataDirEnv = "KAIZEN_DATA_DIR"

const (
	ConfigFileName  = "config.json"
	HabitsFileName  = "habits.json"
	EntriesFileName = "entries.json"
	NotesFileName   = "notes.json"
)

// An unusable KAIZEN_DATA_DIR is an error rather than a fallback, since silently
// writing to the default location puts your history somewhere you will not look.
func (r *FileRepository) DataDir() (string, error) {
	if r.dataDir != "" {
		return r.dataDir, nil
	}

	if raw, set := os.LookupEnv(DataDirEnv); set {
		dir := strings.TrimSpace(raw)
		switch {
		case dir == "":
			return "", fmt.Errorf("%s is set but blank, unset it to use the default location", DataDirEnv)
		case !filepath.IsAbs(dir):
			return "", fmt.Errorf("%s must be an absolute path, got %q", DataDirEnv, dir)
		}
		if info, err := os.Stat(dir); err == nil && !info.IsDir() {
			return "", fmt.Errorf("%s points at %q, which is not a directory", DataDirEnv, dir)
		}
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents", ".kaizen"), nil
}

func (r *FileRepository) DataFilePath(fileName string) (string, error) {
	dir, err := r.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

func (r *FileRepository) EnsureDataDir() (string, error) {
	dir, err := r.DataDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func (r *FileRepository) Exists(fileName string) (bool, error) {
	path, err := r.DataFilePath(fileName)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
