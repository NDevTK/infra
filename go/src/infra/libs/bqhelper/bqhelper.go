// Package bqhelper implements an interface for a bigquery client.  It also
// exposes methods for interacting with a client. It was created to make it
// easier to write unit tests for code that uses bigquery.
package bqhelper

import (
	"errors"
	"net/http"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
)

type ctxKey string

// BQClientKey is used to reference the bq client in the context.
const BQClientKey = ctxKey("bqClient")

// bqClient is an interface for a bigquery client.
type bqClient interface {
	getTable(datasetID string, tableID string) (*bigquery.Table, error)
	getTableMetadata(ctx context.Context, datasetID string,
		tableID string) (*bigquery.TableMetadata, error)
	createTable(ctx context.Context, s bigquery.Schema,
		t *bigquery.Table) error
	updateTable(ctx context.Context, md bigquery.TableMetadataToUpdate,
		t *bigquery.Table) (*bigquery.TableMetadata, error)
}

func getBQClient(ctx context.Context) (bqClient, error) {
	if bq, ok := ctx.Value(BQClientKey).(bqClient); ok && bq != nil {
		return bq, nil
	}
	return nil, errors.New("could not retrieve BigQuery client")
}

// GetTable creates a handle to a BigQuery table in the dataset. To determine
// if a table exists, call GetTableMetadata. If the table does not already
// exist, use CreateTable to create it.
func GetTable(ctx context.Context, datasetID string,
	tableID string) (*bigquery.Table, error) {
	bq, err := getBQClient(ctx)
	if err == nil {
		return bq.getTable(datasetID, tableID)
	}
	return nil, err
}

// GetTableMetadata fetches the metadata for the table.
func GetTableMetadata(ctx context.Context, datasetID string,
	tableID string) (*bigquery.TableMetadata, error) {
	bq, err := getBQClient(ctx)
	if err == nil {
		return bq.getTableMetadata(ctx, datasetID, tableID)
	}
	return nil, err
}

// CreateTable creates a table in the BigQuery service.
func CreateTable(ctx context.Context, s bigquery.Schema,
	t *bigquery.Table) error {
	bq, err := getBQClient(ctx)
	if err == nil {
		return bq.createTable(ctx, s, t)
	}
	return err
}

// UpdateTable modifies specific Table metadata fields.
func UpdateTable(ctx context.Context, md bigquery.TableMetadataToUpdate,
	t *bigquery.Table) (*bigquery.TableMetadata, error) {
	bq, err := getBQClient(ctx)
	if err == nil {
		return bq.updateTable(ctx, md, t)
	}
	return nil, err
}

type prodBQ struct {
	c *bigquery.Client
}

// UseCloudClient adds a prodBQ client as a context value.
func UseCloudClient(ctx context.Context,
	projectID string) (context.Context, error) {
	c, err := bigquery.NewClient(ctx, projectID)
	return context.WithValue(ctx, BQClientKey, c), err
}

func (bq *prodBQ) getTable(datasetID string,
	tableID string) (*bigquery.Table, error) {
	return bq.c.Dataset(datasetID).Table(tableID), nil
}

func (bq *prodBQ) getTableMetadata(ctx context.Context, datasetID string,
	tableID string) (*bigquery.TableMetadata, error) {
	t, err := bq.getTable(datasetID, tableID)
	if err != nil {
		return nil, err
	}
	return t.Metadata(ctx)
}

func (bq *prodBQ) createTable(ctx context.Context, s bigquery.Schema,
	t *bigquery.Table) error {
	return t.Create(ctx, s)
}

func (bq *prodBQ) updateTable(ctx context.Context,
	md bigquery.TableMetadataToUpdate,
	t *bigquery.Table) (*bigquery.TableMetadata, error) {
	return t.Update(ctx, md)
}

type mockBQ struct {
	// map[datasetID+tableID]*bigquery.Table
	tables map[string]*bigquery.Table
	// map[datasetID+tableID]*bigquery.TableMetadata
	tableMetadatas map[string]*bigquery.TableMetadata
}

// UseMemoryClient adds a mockBQ client with maps initialized as a context
// value.
func UseMemoryClient(ctx context.Context) context.Context {
	c := &mockBQ{
		tables:         map[string]*bigquery.Table{},
		tableMetadatas: map[string]*bigquery.TableMetadata{},
	}
	return context.WithValue(ctx, BQClientKey, c)
}

func (bq *mockBQ) getTable(datasetID string,
	tableID string) (*bigquery.Table, error) {
	t, ok := bq.tables[datasetID+tableID]
	if ok {
		return t, nil
	}
	return &bigquery.Table{
		DatasetID: datasetID,
		TableID:   tableID,
	}, nil
}

func (bq *mockBQ) getTableMetadata(ctx context.Context, datasetID string,
	tableID string) (*bigquery.TableMetadata, error) {
	md, ok := bq.tableMetadatas[datasetID+tableID]
	if ok {
		return md, nil
	}
	return nil, &googleapi.Error{Code: http.StatusNotFound}
}

func (bq *mockBQ) createTable(ctx context.Context, s bigquery.Schema,
	t *bigquery.Table) error {
	if t.DatasetID == "" {
		return errors.New("table must have DatasetID")
	}
	if t.TableID == "" {
		return errors.New("table must have TableID")
	}
	md := &bigquery.TableMetadata{
		Schema: s,
		ID:     t.TableID,
	}
	comboKey := t.DatasetID + t.TableID
	bq.tables[comboKey] = t
	bq.tableMetadatas[comboKey] = md
	return nil
}

func (bq *mockBQ) updateTable(ctx context.Context,
	toUpdate bigquery.TableMetadataToUpdate,
	t *bigquery.Table) (*bigquery.TableMetadata, error) {
	if t.DatasetID == "" {
		return nil, errors.New("table must have DatasetID")
	}
	if t.TableID == "" {
		return nil, errors.New("table must have TableID")
	}
	md, ok := bq.tableMetadatas[t.DatasetID+t.TableID]
	if !ok {
		return nil, errors.New("table not in tableMetadatas")
	}
	if toUpdate.Name != nil {
		md.Name = toUpdate.Name.(string)
	}
	if toUpdate.Description != nil {
		md.Description = toUpdate.Description.(string)
	}
	if toUpdate.Schema != nil {
		md.Schema = toUpdate.Schema
	}
	return md, nil
}
