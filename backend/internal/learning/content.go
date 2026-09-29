package learning

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Node struct {
	ID                string   `json:"id"`
	SkillID           string   `json:"skillId"`
	Title             string   `json:"title"`
	Objective         string   `json:"objective"`
	Prompt            string   `json:"prompt"`
	Support           []string `json:"support"`
	AcceptedVariation []string `json:"acceptedVariation"`
	Prerequisites     []string `json:"prerequisites"`
}
type Milestone struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Nodes []Node `json:"nodes"`
}
type Subject struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
}
type Manifest struct {
	SchemaVersion  int         `json:"schemaVersion"`
	ContentVersion string      `json:"contentVersion"`
	Subject        Subject     `json:"subject"`
	ReviewStatus   string      `json:"reviewStatus"`
	Milestones     []Milestone `json:"milestones"`
}
type Catalog map[string]Manifest

type rowQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s Server) manifest(ctx context.Context, q rowQuerier, trackID, subjectID string) (Manifest, error) {
	if m, ok := s.Catalog[subjectID]; ok {
		return m, nil
	}
	var raw []byte
	if err := q.QueryRowContext(ctx, "SELECT manifest FROM learning.track_content WHERE track_id=$1", trackID).Scan(&raw); err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, err
	}
	if m.Subject.ID != subjectID || m.SchemaVersion != 1 {
		return Manifest{}, sql.ErrNoRows
	}
	return m, nil
}

func LoadCatalog(root string) (Catalog, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	catalog := Catalog{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, entry.Name(), "manifest.json"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var manifest Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		if manifest.Subject.ID != entry.Name() || manifest.SchemaVersion != 1 {
			return nil, fmt.Errorf("%s: invalid manifest identity or version", entry.Name())
		}
		catalog[manifest.Subject.ID] = manifest
	}
	if len(catalog) == 0 {
		return nil, fmt.Errorf("no content manifests in %s", root)
	}
	return catalog, nil
}

func (m Manifest) FindNode(id string) (Node, bool) {
	for _, milestone := range m.Milestones {
		for _, node := range milestone.Nodes {
			if node.ID == id {
				return node, true
			}
		}
	}
	return Node{}, false
}

func (m Manifest) Count() int {
	n := 0
	for _, group := range m.Milestones {
		n += len(group.Nodes)
	}
	return n
}
