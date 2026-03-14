package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	baseDir string
}

func NewStore(dir string) *Store {
	return &Store{baseDir: filepath.Join(dir, ".eng-graph", "profiles")}
}

func (s *Store) Create(name string) (*Profile, error) {
	dir := filepath.Join(s.baseDir, name)
	if _, err := os.Stat(dir); err == nil {
		return nil, fmt.Errorf("profile %q already exists", name)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	p := &Profile{Name: name, DisplayName: name}
	return p, s.Save(p)
}

func (s *Store) Load(name string) (*Profile, error) {
	data, err := os.ReadFile(filepath.Join(s.baseDir, name, "profile.json"))
	if err != nil {
		return nil, err
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) Save(p *Profile) error {
	dir := filepath.Join(s.baseDir, p.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "profile.json"), data, 0644)
}

func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func (s *Store) OutputDir(name string) string {
	return filepath.Join(s.baseDir, name, "output")
}

func (s *Store) ProfileDir(name string) string {
	return filepath.Join(s.baseDir, name)
}
