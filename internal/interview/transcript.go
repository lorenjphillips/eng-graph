package interview

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Transcript struct {
	Exchanges []Exchange    `json:"exchanges"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration"`
	Complete  bool          `json:"complete"`
}

func NewTranscript() *Transcript {
	return &Transcript{StartedAt: time.Now()}
}

func (t *Transcript) Add(e Exchange) {
	t.Exchanges = append(t.Exchanges, e)
}

func (t *Transcript) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadTranscript(path string) (*Transcript, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t Transcript
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (t *Transcript) ToJSON() json.RawMessage {
	data, _ := json.MarshalIndent(t, "", "  ")
	return data
}
