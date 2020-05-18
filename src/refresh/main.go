package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ralbertazzi/alfred-bigquery-shortcuts/src"
	"google.golang.org/api/bigquery/v2"

	aw "github.com/deanishe/awgo"
)

const (
	cacheGoogleProjects = "bigquery-cache"
	rateLimit           = time.Second / 100 // 100 calls per second
)

var (
	logger = log.New(os.Stderr, "[refresh] ", log.LstdFlags)
	wf     *aw.Workflow
)

type ProjectDescription struct {
	Name string
	ID   string
}

type DatasetDescription struct {
	ID string
}

type DatasetsResponse struct {
	projectId string
	datasets  []DatasetDescription
	err       error
}

type TableDescription struct {
	ID     string
	Labels map[string]string
}

type TablesResponse struct {
	projectId string
	datasetId string
	tables    []TableDescription
	err       error
}

func FetchGoogleProjects(ctx context.Context) (map[string]ProjectDescription, error) {
	bq_service, err := bigquery.NewService(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery service: %w", err)
	}

	bq_projects_service := bigquery.NewProjectsService(bq_service)

	allProjects := make(map[string]ProjectDescription, 0)

	nextPageToken := ""
	for {
		res, err := bq_projects_service.List().PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		nextPageToken = res.NextPageToken

		for _, project := range res.Projects {
			allProjects[project.Id] = ProjectDescription{
				Name: project.FriendlyName,
				ID:   project.Id,
			}
		}
		if nextPageToken == "" {
			break
		}
	}

	return allProjects, nil
}

func FetchBigQueryDatasets(
	projectId string,
	ctx context.Context,
	throttle <-chan time.Time,
	output chan DatasetsResponse,
) {
	bq_service, err := bigquery.NewService(ctx)
	if err != nil {
		output <- DatasetsResponse{"", nil, fmt.Errorf("failed to create bigquery service: %w", err)}
		return
	}

	bq_dataset_service := bigquery.NewDatasetsService(bq_service)

	allDatasets := make([]DatasetDescription, 0)

	nextPageToken := ""
	for {
		<-throttle
		res, err := bq_dataset_service.List(projectId).PageToken(nextPageToken).Do()
		if err != nil {
			output <- DatasetsResponse{"", nil, err}
			return
		}
		nextPageToken = res.NextPageToken

		for _, dataset := range res.Datasets {
			allDatasets = append(allDatasets, DatasetDescription{
				ID: dataset.DatasetReference.DatasetId,
			})
		}
		if nextPageToken == "" {
			break
		}
	}

	output <- DatasetsResponse{projectId, allDatasets, nil}
}

func FetchBigQueryTables(
	projectId string,
	datasetId string,
	maxTablesPerDataset int,
	ctx context.Context,
	throttle <-chan time.Time,
	output chan TablesResponse,
) {
	bq_service, err := bigquery.NewService(ctx)
	if err != nil {
		output <- TablesResponse{"", "", nil, fmt.Errorf("failed to create bigquery service: %w", err)}
		return
	}

	bq_tables_service := bigquery.NewTablesService(bq_service)

	allTables := make([]TableDescription, 0)

	maxResults := 1000

	for i, nextPageToken := 0, ""; i < maxTablesPerDataset/maxResults; i++ {
		<-throttle
		res, err := bq_tables_service.List(projectId, datasetId).MaxResults(int64(maxResults)).PageToken(nextPageToken).Do()
		if err != nil {
			output <- TablesResponse{"", "", nil, err}
			return
		}
		nextPageToken = res.NextPageToken

		for _, table := range res.Tables {
			allTables = append(allTables, TableDescription{
				ID: table.TableReference.TableId, Labels: table.Labels,
			})
		}
		if nextPageToken == "" {
			break
		}
	}

	output <- TablesResponse{projectId, datasetId, allTables, nil}
}

func init() {
	wf = aw.New()
	wf.Configure(aw.TextErrors(true))
}

func run() {
	var maxTablesPerDataset int
	flag.IntVar(&maxTablesPerDataset, "max_tables_per_dataset", 1000, "maximum number of fetched tables per dataset")
	flag.Parse()

	ctx := context.Background()

	logger.Printf("Refreshing projects and datasets")
	projects, err := FetchGoogleProjects(ctx)
	if err != nil {
		wf.FatalError(err)
	}

	data := make([]src.ProjectEntry, 0)

	throttle := time.Tick(rateLimit)

	datasetsChan := make(chan DatasetsResponse)
	n_dataset_responses := 0
	for projectId := range projects {
		n_dataset_responses += 1
		go FetchBigQueryDatasets(projectId, ctx, throttle, datasetsChan)
	}

	projectsAndDatasets := make(map[string][]DatasetDescription, 0)

	for i := 0; i < n_dataset_responses; i++ {
		datasetsReponse := <-datasetsChan
		if datasetsReponse.err != nil {
			logger.Printf("Could not retrieve  datasets for project %s: %w", datasetsReponse.projectId, datasetsReponse.err)
		} else {
			logger.Printf("Retrieved %d datasets from project %s", len(datasetsReponse.datasets), datasetsReponse.projectId)
			projectsAndDatasets[datasetsReponse.projectId] = datasetsReponse.datasets
		}
	}

	tablesChan := make(chan TablesResponse)
	n_table_responses := 0
	for projectId, datasets := range projectsAndDatasets {
		for _, dataset := range datasets {
			n_table_responses += 1
			go FetchBigQueryTables(projectId, dataset.ID, maxTablesPerDataset, ctx, throttle, tablesChan)
		}
	}

	datasetsAndTables := make(map[string]map[string][]TableDescription, 0)
	for projectId, _ := range projectsAndDatasets {
		datasetsAndTables[projectId] = make(map[string][]TableDescription, 0)
	}

	for i := 0; i < n_table_responses; i++ {
		tablesResponse := <-tablesChan
		if tablesResponse.err != nil {
			logger.Printf("Could not retrieve tables for project %s dataset %s: %w", tablesResponse.projectId, tablesResponse.datasetId, tablesResponse.err)
		} else {
			logger.Printf("Retrieved %d tables from project %s dataset %s", len(tablesResponse.tables), tablesResponse.projectId, tablesResponse.datasetId)
			datasetsAndTables[tablesResponse.projectId][tablesResponse.datasetId] = tablesResponse.tables
		}
	}

	for projectId := range datasetsAndTables {
		datasetEntries := make([]src.DatasetEntry, 0)
		for datasetId, tables := range datasetsAndTables[projectId] {
			tableEntries := make([]src.TableEntry, 0)

			for _, table := range tables {
				tableEntries = append(tableEntries, src.TableEntry{TableId: table.ID, Labels: table.Labels})
			}

			datasetEntries = append(datasetEntries, src.DatasetEntry{DatasetId: datasetId, Tables: tableEntries})
		}

		data = append(data, src.ProjectEntry{
			ProjectId:   projectId,
			ProjectName: projects[projectId].Name,
			Datasets:    datasetEntries,
		})
	}

	err = src.StoreBigQueryData(wf, data)
	if err != nil {
		wf.FatalError(err)
	}
	logger.Printf("Refresh done")
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
