package src

import (
	"fmt"
	"os"
	"path"

	aw "github.com/deanishe/awgo"
)

const (
	projectsCacheFile = "projects.json"
)

type TableEntry struct {
	TableId string            `json:"tableId"`
	Labels  map[string]string `json:"labels"`
}

type DatasetEntry struct {
	DatasetId string       `json:"datasetId"`
	Tables    []TableEntry `json:"tables"`
}

type ProjectEntry struct {
	ProjectId   string         `json:"projectId"`
	ProjectName string         `json:"projectName"`
	Datasets    []DatasetEntry `json:"datasets"`
}

type ProjectEntrySimple struct {
	ProjectId   string `json:"projectId"`
	ProjectName string `json:"projectName"`
}

func projectDir(wf *aw.Workflow) string {
	return path.Join(wf.DataDir(), "projects")
}

func projectCachePath(wf *aw.Workflow, projectId string) string {
	return path.Join("projects", projectId)
}

func StoreBigQueryData(wf *aw.Workflow, data []ProjectEntry) error {
	projectEntriesSimple := make([]ProjectEntrySimple, 0)
	for _, projectEntry := range data {
		projectEntriesSimple = append(projectEntriesSimple, ProjectEntrySimple{ProjectId: projectEntry.ProjectId, ProjectName: projectEntry.ProjectName})
	}
	if err := wf.Data.StoreJSON(projectsCacheFile, projectEntriesSimple); err != nil {
		return fmt.Errorf("error storing projects: %w", err)
	}

	if err := os.RemoveAll(projectDir(wf)); err != nil {
		return fmt.Errorf("error while removing projects dir: %w", err)
	}

	if err := os.MkdirAll(projectDir(wf), 0700); err != nil {
		return fmt.Errorf("error while creating projects dir: %w", err)
	}

	for _, projectEntry := range data {
		if err := wf.Data.StoreJSON(projectCachePath(wf, projectEntry.ProjectId), projectEntry); err != nil {
			return fmt.Errorf("error storing data for project %s: %w", projectEntry.ProjectId, err)
		}
	}

	return nil
}

func LoadProjects(wf *aw.Workflow) ([]ProjectEntrySimple, error) {
	var projects []ProjectEntrySimple
	if err := wf.Data.LoadJSON(projectsCacheFile, &projects); err != nil {
		return nil, fmt.Errorf("error loading projects: %w", err)
	}

	return projects, nil
}

func LoadProjectEntry(wf *aw.Workflow, projectId string) (*ProjectEntry, error) {
	projectEntry := &ProjectEntry{}
	if err := wf.Data.LoadJSON(projectCachePath(wf, projectId), projectEntry); err != nil {
		return nil, fmt.Errorf("error loading datasets and table for project %s: %w", projectId, err)
	}

	return projectEntry, nil
}
