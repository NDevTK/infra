package bqhelper

import (
	"net/http"
	"reflect"
	"testing"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
)

type testSchemaOne struct {
	FieldOne int
}

type testSchemaTwo struct {
	FieldOne string
}

func createContext() context.Context {
	return UseMemoryClient(context.Background())
}

func getTestTable(ctx context.Context, datasetID string, tableID string,
	t *testing.T) *bigquery.Table {
	table, err := GetTable(ctx, datasetID, tableID)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	return table
}

func createTestTable(ctx context.Context, t *testing.T, datasetID string,
	tableID string, s bigquery.Schema,
	toUpdate bigquery.TableMetadataToUpdate) *bigquery.Table {
	table := getTestTable(ctx, datasetID, tableID, t)
	err := CreateTable(ctx, s, table)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	table, err = GetTable(ctx, datasetID, tableID)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	md, err := GetTableMetadata(ctx, datasetID, tableID)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	if toUpdate.Name != nil {
		md.Name = toUpdate.Name.(string)
	}
	if toUpdate.Description != nil {
		md.Description = toUpdate.Description.(string)
	}
	return table
}

func tableWasCreated(ctx context.Context, datasetID string,
	tableID string) bool {
	_, err := GetTableMetadata(ctx, datasetID, tableID)
	if err == nil {
		return true
	}
	return false
}

func createTestTableWithMetadata(ctx context.Context, name string, desc string,
	t *testing.T, datasetID string, tableID string,
	s bigquery.Schema) *bigquery.Table {
	toUpdate := bigquery.TableMetadataToUpdate{
		Name:        name,
		Description: desc,
	}
	table := createTestTable(ctx, t, datasetID, tableID, s,
		toUpdate)
	return table
}

func TestGetTable(t *testing.T) {
	datasetID := "test_dataset"
	tableID := "test_table_1"
	want := &bigquery.Table{
		DatasetID: datasetID,
		TableID:   tableID,
	}
	ctx := createContext()
	t.Run("WithoutCreate", func(t *testing.T) {
		have := getTestTable(ctx, datasetID, tableID, t)
		if !reflect.DeepEqual(have, want) {
			t.Logf("Have: %v Want: %v", have, want)
			t.Fail()
		}
	})
	t.Run("WithCreate", func(t *testing.T) {
		have := getTestTable(ctx, datasetID, tableID, t)
		s, err := bigquery.InferSchema(testSchemaOne{})
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		err = CreateTable(ctx, s, have)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		have, err = GetTable(ctx, datasetID, tableID)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		if !reflect.DeepEqual(have, want) {
			t.Logf("Have: %v Want: %v", have, want)
			t.Fail()
		}
		if !tableWasCreated(ctx, datasetID, tableID) {
			t.Log("Have: table not created")
			t.Log("Want: table created")
			t.Fail()
		}
	})
}

func TestGetTableMetadata(t *testing.T) {
	ctx := createContext()
	datasetID := "test_dataset"
	tableID := "test_table_1"
	t.Run("WithoutCreate", func(t *testing.T) {
		_, err := GetTableMetadata(ctx, datasetID, tableID)
		e, ok := err.(*googleapi.Error)
		if !(ok && e.Code == http.StatusNotFound) {
			t.Fail()
		}
	})
	t.Run("WithCreate", func(t *testing.T) {
		s, err := bigquery.InferSchema(testSchemaOne{})
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		_ = createTestTable(ctx, t, datasetID, tableID, s,
			bigquery.TableMetadataToUpdate{})
		md, err := GetTableMetadata(ctx, datasetID, tableID)
		want := &bigquery.TableMetadata{
			ID:     tableID,
			Schema: s,
		}
		if !(reflect.DeepEqual(md, want)) {
			t.Logf("Have: %v Want: %v", md, want)
			t.Fail()
		}
	})
}

func TestCreateTable(t *testing.T) {
	ctx := createContext()
	datasetID := "test_dataset"
	tableID := "test_table_1"
	s, err := bigquery.InferSchema(testSchemaOne{})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Run("WithoutDatasetID", func(t *testing.T) {
		table := getTestTable(ctx, "", tableID, t)
		err := CreateTable(ctx, s, table)
		if err == nil {
			t.Log("Have: no error")
			t.Log("Want: error")
			t.Fail()
		}
		if tableWasCreated(ctx, datasetID, tableID) {
			t.Log("Have: table created")
			t.Log("Want: table not created")
			t.Fail()
		}
	})
	t.Run("WithoutTableID", func(t *testing.T) {
		table := getTestTable(ctx, datasetID, "", t)
		err := CreateTable(ctx, s, table)
		if err == nil {
			t.Log("Have: no error")
			t.Log("Want: error")
			t.Fail()
		}
		if tableWasCreated(ctx, datasetID, tableID) {
			t.Log("Have: table created")
			t.Log("Want: table not created")
			t.Fail()
		}
	})
	t.Run("Successful", func(t *testing.T) {
		table := getTestTable(ctx, datasetID, tableID, t)
		err := CreateTable(ctx, s, table)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		if !tableWasCreated(ctx, datasetID, tableID) {
			t.Log("Have: table not created")
			t.Log("Want: table created")
			t.Fail()
		}
	})
}

func TestUpdateTable(t *testing.T) {
	ctx := createContext()
	datasetID := "test_dataset"
	tableID := "test_table_1"
	s, err := bigquery.InferSchema(testSchemaOne{})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Run("WithoutDatasetID", func(t *testing.T) {
		table := getTestTable(ctx, "", tableID, t)
		toUpdate := bigquery.TableMetadataToUpdate{}
		_, err := UpdateTable(ctx, toUpdate, table)
		if err == nil {
			t.Fail()
		}
	})
	t.Run("WithoutTableID", func(t *testing.T) {
		table := getTestTable(ctx, datasetID, "", t)
		toUpdate := bigquery.TableMetadataToUpdate{}
		_, err := UpdateTable(ctx, toUpdate, table)
		if err == nil {
			t.Fail()
		}
	})
	t.Run("TableNotCreated", func(t *testing.T) {
		table := getTestTable(ctx, datasetID, tableID, t)
		toUpdate := bigquery.TableMetadataToUpdate{}
		_, err := UpdateTable(ctx, toUpdate, table)
		if err == nil {
			t.Fail()
		}
	})
	t.Run("UpdateNoop", func(t *testing.T) {
		name := "test name"
		desc := "test desc"
		table := createTestTableWithMetadata(ctx, name, desc, t,
			datasetID, tableID, s)
		toUpdate := bigquery.TableMetadataToUpdate{}
		md, err := UpdateTable(ctx, toUpdate, table)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		if md.Name != name || md.Description != desc ||
			!reflect.DeepEqual(md.Schema, s) {
			t.Logf("Have: %v %v %v", md.Name, md.Description,
				md.Schema)
			t.Logf("Want: %v %v %v", name, desc, s)
			t.Fail()
		}
	})
	t.Run("UpdateOneThing", func(t *testing.T) {
		name := "test name"
		desc := "test desc"
		table := createTestTableWithMetadata(ctx, name, desc, t,
			datasetID, tableID, s)
		newName := "new test name"
		toUpdate := bigquery.TableMetadataToUpdate{Name: newName}
		md, err := UpdateTable(ctx, toUpdate, table)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		if md.Name != newName {
			t.Logf("Have: %v", md.Name)
			t.Logf("Want: %v", newName)
			t.Fail()
		}
		if md.Description != desc ||
			!reflect.DeepEqual(md.Schema, s) {
			t.Logf("Have: %v %v", md.Description, md.Schema)
			t.Logf("Want: %v %v", desc, s)
			t.Fail()
		}
	})
	t.Run("UpdateEverything", func(t *testing.T) {
		name := "test name"
		desc := "test desc"
		table := createTestTableWithMetadata(ctx, name, desc, t,
			datasetID, tableID, s)
		newName := "new test name"
		newDesc := "new description"
		newS, err := bigquery.InferSchema(testSchemaTwo{})
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		toUpdate := bigquery.TableMetadataToUpdate{
			Name:        newName,
			Description: newDesc,
			Schema:      newS,
		}
		md, err := UpdateTable(ctx, toUpdate, table)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		if md.Name != newName || md.Description != newDesc ||
			!reflect.DeepEqual(md.Schema, newS) {
			t.Logf("Have: %v %v %v", md.Name, md.Description,
				md.Schema)
			t.Logf("Want: %v %v %v", newName, newDesc, newS)
			t.Fail()
		}
	})
}
