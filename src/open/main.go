package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/deanishe/awgo/fuzzy"
	"github.com/ralbertazzi/alfred-bigquery-shortcuts/src"

	aw "github.com/deanishe/awgo"
)

const (
	projectUrlTemplate = "https://console.cloud.google.com/bigquery?project={project}&p={project}"
	datasetUrlTemplate = "https://console.cloud.google.com/bigquery?project={project}&p={project}&d={dataset}&page=dataset"
	tableUrlTemplate   = "https://console.cloud.google.com/bigquery?project={project}&p={project}&d={dataset}&t={table}&page=table"
)

var (
	logger = log.New(os.Stderr, "[open] ", log.LstdFlags)
	wf     *aw.Workflow

	query   string
	project string
	dataset string
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

func init() {
	sopts := []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(aw.SortOptions(sopts...))
	wf.Configure(aw.TextErrors(true))

	flag.StringVar(&query, "query", "", "search query")
	flag.StringVar(&project, "project", "", "google cloud project")
	flag.StringVar(&dataset, "dataset", "", "bigquery dataset")
}

func run() {

	logger.Print("Hello there")
	flag.Parse()

	logger.Printf("Project is %s", project)
	logger.Printf("Dataset is %s", dataset)
	logger.Printf("Query is %s", query)

	if project == "" {
		projectEntries, err := src.LoadProjects(wf)
		if err != nil {
			wf.FatalError(err)
			return
		}

		for _, p := range projectEntries {
			projectUrl := strings.NewReplacer("{project}", p.ProjectId).Replace(projectUrlTemplate)
			wf.NewItem(p.ProjectName).Var("project", p.ProjectId).Var("projectUrl", projectUrl).Subtitle(p.ProjectId).UID(p.ProjectId).Valid(true)
		}
	} else {
		p, err := src.LoadProjectEntry(wf, project)
		if err != nil {
			wf.FatalError(err)
			return
		}

		if p.ProjectId == project {
			for _, d := range p.Datasets {
				if dataset == "" {
					datasetUrl := strings.NewReplacer("{project}", p.ProjectId, "{dataset}", d.DatasetId).Replace(datasetUrlTemplate)
					wf.NewItem(d.DatasetId).Var("project", p.ProjectId).Var("dataset", d.DatasetId).Var("datasetUrl", datasetUrl).UID(d.DatasetId).Valid(true)
				} else {
					if d.DatasetId == dataset {
						for _, t := range d.Tables {
							tableUrl := strings.NewReplacer("{project}", p.ProjectId, "{dataset}", d.DatasetId, "{table}", t.TableId).Replace(tableUrlTemplate)

							tableMatch := strings.Replace(t.TableId, "_", " ", -1)
							subTitle := t.TableId
							if len(t.Labels) > 0 {
								labels := make([]string, 0)
								subtitleParts := make([]string, 0)

								for labelKey, labelValue := range t.Labels {
									labels = append(labels, labelValue)
									subtitleParts = append(subtitleParts, fmt.Sprintf("%s: %s", labelKey, labelValue))
								}
								subTitle = strings.Join(subtitleParts, ", ")
								tableMatch = strings.Join(labels, " ") + " " + tableMatch + " " + subTitle
							}

							wf.NewItem(t.TableId).Subtitle(subTitle).Match(tableMatch).Var("tableUrl", tableUrl).UID(t.TableId).Valid(true)
						}
					}
				}
			}
		}
	}

	if query != "" {
		wf.Filter(query)
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
