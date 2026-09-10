package changeset

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

const Dir = ".changesets"

type Changeset struct {
	ID          string   `yaml:"id"`
	Features    []string `yaml:"features"`
	Targets     []string `yaml:"targets"`
	Type        string   `yaml:"type"`
	Bump        string   `yaml:"bump"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Date        string   `yaml:"date"`
}

func NewID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d-%x", time.Now().Unix(), b)
}

func Write(cs *Changeset) (string, error) {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return "", fmt.Errorf("error creating changesets directory: %w", err)
	}

	data, err := yaml.Marshal(cs)
	if err != nil {
		return "", fmt.Errorf("error marshalling changeset: %w", err)
	}

	p := filepath.Join(Dir, cs.ID+".yaml")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return "", fmt.Errorf("error writing changeset: %w", err)
	}

	return p, nil
}

func ReadAll() ([]*Changeset, error) {
	entries, err := os.ReadDir(Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading changesets directory: %w", err)
	}

	changesets := make([]*Changeset, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(Dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("error reading changeset %s: %w", entry.Name(), err)
		}

		var cs Changeset
		if err := yaml.Unmarshal(data, &cs); err != nil {
			return nil, fmt.Errorf("error parsing changeset %s: %w", entry.Name(), err)
		}

		changesets = append(changesets, &cs)
	}

	// Sort by date then ID for deterministic ordering
	sort.Slice(changesets, func(i, j int) bool {
		if changesets[i].Date != changesets[j].Date {
			return changesets[i].Date < changesets[j].Date
		}
		return changesets[i].ID < changesets[j].ID
	})

	return changesets, nil
}

func DeleteAll() error {
	entries, err := os.ReadDir(Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error reading changesets directory: %w", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		if err := os.Remove(filepath.Join(Dir, entry.Name())); err != nil {
			return fmt.Errorf("error deleting changeset %s: %w", entry.Name(), err)
		}
	}

	return nil
}
