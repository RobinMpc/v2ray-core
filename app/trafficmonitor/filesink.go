package trafficmonitor

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// FileSink writes traffic metrics as JSON lines to a file.
type FileSink struct {
	file   *os.File
	enc    *json.Encoder
}

// NewFileSink creates a new FileSink. Creates parent directories if needed.
func NewFileSink(path string) (*FileSink, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileSink{file: f, enc: json.NewEncoder(f)}, nil
}

func (s *FileSink) Write(snapshots []MetricsSnapshot) error {
	if s == nil || s.file == nil {
		return nil
	}
	for _, snap := range snapshots {
		if err := s.enc.Encode(snap); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileSink) Close() error {
	if s == nil || s.file == nil {
		return nil
	}
	return s.file.Close()
}
