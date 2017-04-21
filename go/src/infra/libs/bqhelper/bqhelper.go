package main

import (
	"errors"
	"net/http"

	cloudBQ "cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
)

type ctxKey string

// BQClientKey is used to reference the bq client in the context.
const BQClientKey = ctxKey("bqClient")

// biquery is a  wrapper for a bigquery client.
type bigquery interface {
	getTable(datasetID string, tableID string) (*cloudBQ.Table, error)
	getTableMetadata(ctx context.Context, datasetID string,
		tableID string) (*cloudBQ.TableMetadata, error)
	createTable(ctx context.Context, s cloudBQ.Schema,
		t *cloudBQ.Table) error
	updateTable(ctx context.Context, md cloudBQ.TableMetadataToUpdate,
		t *cloudBQ.Table) (*cloudBQ.TableMetadata, error)
}

func bq(ctx context.Context) (bigquery, error) {
	if bq, ok := ctx.Value(BQClientKey).(bigquery); ok && bq != nil {
		return bq, nil
	}
	return nil, errors.New("Could not retrieve BigQuery client")
}

// GetTable creates a handle to a BigQuery table in the dataset. To determine
// if a table exists, call GetTableMetadata. If the table does not already
// exist, use CreateTable to create it.
func GetTable(ctx context.Context, datasetID string,
	tableID string) (*cloudBQ.Table, error) {
	bq, err := bq(ctx)
	if err == nil {
		return bq.getTable(datasetID, tableID)
	}
	return nil, err
}

// GetTableMetadata fetches the metadata for the table.
func GetTableMetadata(ctx context.Context, datasetID string,
	tableID string) (*cloudBQ.TableMetadata, error) {
	bq, err := bq(ctx)
	if err == nil {
		return bq.getTableMetadata(ctx, datasetID, tableID)
	}
	return nil, err
}

// CreateTable creates a table in the BigQuery service.
func CreateTable(ctx context.Context, s cloudBQ.Schema,
	t *cloudBQ.Table) error {
	bq, err := bq(ctx)
	if err == nil {
		return bq.createTable(ctx, s, t)
	}
	return err
}

// UpdateTable modifies specific Table metadata fields.
func UpdateTable(ctx context.Context, md cloudBQ.TableMetadataToUpdate,
	t *cloudBQ.Table) (*cloudBQ.TableMetadata, error) {
	bq, err := bq(ctx)
	if err == nil {
		return bq.updateTable(ctx, md, t)
	}
	return nil, err
}

// ProdBQ is a wrapped cloud bigquery client.
type ProdBQ struct {
	c *cloudBQ.Client
}

func (bq *ProdBQ) getTable(datasetID string,
	tableID string) (*cloudBQ.Table, error) {
	return bq.c.Dataset(datasetID).Table(tableID), nil
}

func (bq *ProdBQ) getTableMetadata(ctx context.Context, datasetID string,
	tableID string) (*cloudBQ.TableMetadata, error) {
	t, err := bq.getTable(datasetID, tableID)
	if err != nil {
		return nil, err
	}
	return t.Metadata(ctx)
}

func (bq *ProdBQ) createTable(ctx context.Context, s cloudBQ.Schema,
	t *cloudBQ.Table) error {
	return t.Create(ctx, s)
}

func (bq *ProdBQ) updateTable(ctx context.Context,
	md cloudBQ.TableMetadataToUpdate,
	t *cloudBQ.Table) (*cloudBQ.TableMetadata, error) {
	return t.Update(ctx, md)
}

// MockBQ is a local bigquery client mock.
type MockBQ struct {
	// map[datasetID+tableID]*cloudBQ.Table
	tables map[string]*cloudBQ.Table
	// map[datasetID+tableID]*cloudBQ.TableMetadata
	tableMetadatas map[string]*cloudBQ.TableMetadata
}

// NewMockClient returns a MockBQ pointer with maps initialized.
func NewMockClient() *MockBQ {
	return &MockBQ{
		tables:         map[string]*cloudBQ.Table{},
		tableMetadatas: map[string]*cloudBQ.TableMetadata{},
	}
}

func (bq *MockBQ) getTable(datasetID string,
	tableID string) (*cloudBQ.Table, error) {
	t, ok := bq.tables[datasetID+tableID]
	if ok {
		return t, nil
	}
	return &cloudBQ.Table{
		DatasetID: datasetID,
		TableID:   tableID,
	}, nil
}

func (bq *MockBQ) getTableMetadata(ctx context.Context, datasetID string,
	tableID string) (*cloudBQ.TableMetadata, error) {
	md, ok := bq.tableMetadatas[datasetID+tableID]
	if ok {
		return md, nil
	}
	return nil, &googleapi.Error{Code: http.StatusNotFound}
}

func (bq *MockBQ) createTable(ctx context.Context, s cloudBQ.Schema,
	t *cloudBQ.Table) error {
	if t.DatasetID == "" {
		return errors.New("Table must have DatasetID")
	}
	if t.TableID == "" {
		return errors.New("Table must have TableID")
	}
	md := &cloudBQ.TableMetadata{
		Schema: s,
		ID:     t.TableID,
	}
	comboKey := t.DatasetID + t.TableID
	bq.tables[comboKey] = t
	bq.tableMetadatas[comboKey] = md
	return nil
}

func (bq *MockBQ) updateTable(ctx context.Context,
	toUpdate cloudBQ.TableMetadataToUpdate,
	t *cloudBQ.Table) (*cloudBQ.TableMetadata, error) {
	if t.DatasetID == "" {
		return nil, errors.New("Table must have DatasetID")
	}
	if t.TableID == "" {
		return nil, errors.New("Table must have TableID")
	}
	md, ok := bq.tableMetadatas[t.DatasetID+t.TableID]
	if !ok {
		return nil, errors.New("Table not in tableMetadatas")
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
