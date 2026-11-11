package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/dlvhdr/gh-dash/v4/internal/recovery"
)

// prepareRecovery returns no store for an ordinary terminal or for a resume
// with no interrupted Dash. That makes --resume safe in a shell startup hook.
func prepareRecovery(path string, resume bool) (*recovery.Store, recovery.Checkpoint, error) {
	var err error
	if path == "" {
		path, err = recovery.PanePath()
	}
	if err != nil || path == "" {
		return nil, recovery.Checkpoint{}, err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, recovery.Checkpoint{}, err
	}
	store, err := recovery.Open(path)
	if err != nil {
		if resume && errors.Is(err, recovery.ErrBusy) {
			return nil, recovery.Checkpoint{}, nil
		}
		return nil, recovery.Checkpoint{}, err
	}
	if resume {
		saved, err := recovery.Load(path)
		if errors.Is(err, os.ErrNotExist) || (err == nil && !saved.Active) {
			store.Close()
			return nil, recovery.Checkpoint{}, nil
		}
		if err != nil {
			store.Close()
			return nil, recovery.Checkpoint{}, err
		}
		if saved.Config != "" {
			if _, err := os.Stat(saved.Config); err != nil {
				store.Close()
				return nil, recovery.Checkpoint{}, err
			}
		}
		if err := os.Chdir(saved.Cwd); err != nil {
			store.Close()
			return nil, recovery.Checkpoint{}, err
		}
		cfgFlag = saved.Config
		return store, saved, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		store.Close()
		return nil, recovery.Checkpoint{}, err
	}
	cfg := cfgFlag
	if cfg == "" {
		cfg = os.Getenv("GH_DASH_CONFIG")
	}
	if cfg != "" {
		cfg, err = filepath.Abs(cfg)
		if err != nil {
			store.Close()
			return nil, recovery.Checkpoint{}, err
		}
	}
	return store, recovery.Checkpoint{Cwd: cwd, Config: cfg}, nil
}
