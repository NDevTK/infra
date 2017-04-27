package main

import (
	"errors"
	"net/http"

	cloudBQ "cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
)

// Separate package

func bq(ctx context.Context) (bigquery, error) {
	if bq, ok := ctx.Value(bqClientKey).(bigquery); ok && bq != nil {
		return bq, nil
	}
	return nil, errors.New("Could not retrieve BigQuery client")
}

func getTable(ctx context.Context, datasetID string,
		tableID string) (*cloudBQ.Table, error) {
	bq, err := bq(ctx)
	if err == nil {
		return bq.getTable(datasetID, tableID)
	}
	return nil, err
}

func getTableMetadata(ctx context.Context, datasetID string,
		tableID string) (*cloudBQ.TableMetadata, error) {
	bq, err := bq(ctx)
	if err == nil {
		return bq.getTableMetadata(ctx, datasetID, tableID)
	}
	return nil, err
}

func createTable(ctx context.Context, s cloudBQ.Schema,
		t *cloudBQ.Table) error {
	bq, err := bq(ctx)
	if err == nil {
		return bq.createTable(ctx, s, t)
	}
	return err
}

func updateTable(ctx context.Context, md cloudBQ.TableMetadataToUpdate,
		t *cloudBQ.Table) (*cloudBQ.TableMetadata, error) {
	bq, err := bq(ctx)
	if err == nil {
		return bq.updateTable(ctx, md, t)
	}
	return nil, err
}

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

type prodBQ struct {
	c *cloudBQ.Client
}

func (bq *prodBQ) getTable(datasetID string,
		tableID string) (*cloudBQ.Table, error) {
	return bq.c.Dataset(datasetID).Table(tableID), nil
}

func (bq *prodBQ) getTableMetadata(ctx context.Context, datasetID string,
		tableID string) (*cloudBQ.TableMetadata, error) {
	t, err := bq.getTable(datasetID, tableID)
	if err != nil {
		return nil, err
	}
	return t.Metadata(ctx)
}

func (bq *prodBQ) createTable(ctx context.Context, s cloudBQ.Schema,
		t *cloudBQ.Table) error {
	return t.Create(ctx, s)
}

func (bq *prodBQ) updateTable(ctx context.Context,
		md cloudBQ.TableMetadataToUpdate,
		t *cloudBQ.Table) (*cloudBQ.TableMetadata, error) {
	return t.Update(ctx, md)
}

type mockBQ struct {
	// map[datasetID+tableID]*cloudBQ.Table
	tables map[string]*cloudBQ.Table
	// map[datasetID+tableID]*cloudBQ.TableMetadata
	tableMetadatas map[string]*cloudBQ.TableMetadata
}

func newMockClient() *mockBQ {
	return &mockBQ{
		tables:         map[string]*cloudBQ.Table{},
		tableMetadatas: map[string]*cloudBQ.TableMetadata{},
	}
}

func (bq *mockBQ) getTable(datasetID string,
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

func (bq *mockBQ) getTableMetadata(ctx context.Context, datasetID string,
		tableID string) (*cloudBQ.TableMetadata, error) {
	md, ok := bq.tableMetadatas[datasetID+tableID]
	if ok {
		return md, nil
	}
	return nil, &googleapi.Error{Code: http.StatusNotFound}
}

func (bq *mockBQ) createTable(ctx context.Context, s cloudBQ.Schema,
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

func (bq *mockBQ) updateTable(ctx context.Context,
		toUpdate cloudBQ.TableMetadataToUpdate,
		t *cloudBQ.Table) (*cloudBQ.TableMetadata, error) {
	if t.DatasetID == "" {
		return nil, errors.New("Table must have DatasetID")
	}
	if t.TableID == "" {
		return nil, errors.New("Table must have TableID")
	}
	md := bq.tableMetadatas[t.DatasetID + t.TableID]
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
