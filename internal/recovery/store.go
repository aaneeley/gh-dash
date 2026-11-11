// Package recovery stores pane-specific checkpoints. Checkpoints contain UI
// navigation only, never credentials, GitHub response bodies or commands to run.
package recovery

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const Version = 1
const maxCheckpointSize = 1 << 20

var ErrBusy = errors.New("dash is already running in this pane")

type Checkpoint struct {
	Version int             `json:"version"`
	Active  bool            `json:"active"`
	Cwd     string          `json:"cwd"`
	Config  string          `json:"config,omitempty"`
	UI      json.RawMessage `json:"ui"`
}

// PanePath returns no path outside Rex. A different pane gets a different file,
// even when both panes are looking at the same repository.
func PanePath() (string, error) {
	block := os.Getenv("REX_BLOCK")
	if block == "" {
		return "", nil
	}
	root := os.Getenv("GH_DASH_STATE_DIR")
	if root == "" {
		root = os.Getenv("XDG_STATE_HOME")
		if root == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			root = filepath.Join(home, ".local", "state")
		}
		root = filepath.Join(root, "gh-dash", "rex")
	}
	key := sha256.Sum256([]byte(block))
	return filepath.Join(root, hex.EncodeToString(key[:])+".json"), nil
}

func Load(path string) (Checkpoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return Checkpoint{}, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxCheckpointSize+1))
	if err != nil {
		return Checkpoint{}, err
	}
	if len(b) > maxCheckpointSize {
		return Checkpoint{}, errors.New("checkpoint is too large")
	}
	var c Checkpoint
	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	if c.Version != Version {
		return c, fmt.Errorf("unsupported checkpoint version %d", c.Version)
	}
	if !filepath.IsAbs(c.Cwd) || (c.Config != "" && !filepath.IsAbs(c.Config)) {
		return c, errors.New("checkpoint paths must be absolute")
	}
	return c, nil
}

type Store struct {
	path string
	lock *os.File
	last []byte
}

// Open takes a process-lifetime lock. A nested shell must not resume a second
// Dash over the first one, and two processes must not overwrite each other.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, err
	}
	return &Store{path: path, lock: f}, nil
}

func (s *Store) Close() error { return s.lock.Close() }

// Save only writes when navigation changed, not on spinner or refresh frames.
// The rename prevents partial JSON after a crash; Sync makes checkpoints
// durable instead of depending on a clean shutdown.
func (s *Store) Save(c Checkpoint) error {
	c.Version = Version
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if len(b) > maxCheckpointSize {
		return errors.New("checkpoint is too large")
	}
	if bytes.Equal(b, s.last) {
		return nil
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".checkpoint-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err = f.Write(b); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	if dir, err := os.Open(filepath.Dir(s.path)); err == nil {
		// Some filesystems do not support directory Sync. The checkpoint
		// itself was already synced, so that is not a save failure.
		_ = dir.Sync()
		_ = dir.Close()
	}
	s.last = b
	return nil
}
