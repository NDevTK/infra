package main

import (
	"net/http"
	"reflect"
	"testing"

	cloudBQ "cloud.google.com/go/bigquery"
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
	return context.WithValue(context.Background(), BQClientKey,
		NewMockClient())
}

func getTestTable(ctx context.Context, datasetID string, tableID string,
	t *testing.T) *cloudBQ.Table {
	table, err := GetTable(ctx, datasetID, tableID)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	return table
}

func createTestTable(ctx context.Context, t *testing.T, datasetID string,
	tableID string, s cloudBQ.Schema,
	toUpdate cloudBQ.TableMetadataToUpdate) *cloudBQ.Table {
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
	s cloudBQ.Schema) *cloudBQ.Table {
	toUpdate := cloudBQ.TableMetadataToUpdate{
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
	want := &cloudBQ.Table{
		DatasetID: datasetID,
		TableID:   tableID,
	}
	ctx := createContext()
	t.Run("WithoutCreate", func(t *testing.T) {
		have := getTestTable(ctx, datasetID, tableID, t)
		if !reflect.DeepEqual(have, want) {
			t.Logf("Want: %v Have: %v", want, have)
			t.Fail()
		}
	})
	t.Run("WithCreate", func(t *testing.T) {
		have := getTestTable(ctx, datasetID, tableID, t)
		s, err := cloudBQ.InferSchema(testSchemaOne{})
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
			t.Logf("Want: %v Have: %v", want, have)
			t.Fail()
		}
		if !tableWasCreated(ctx, datasetID, tableID) {
			t.Log("Want: table created")
			t.Log("Have: table not created")
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
		s, err := cloudBQ.InferSchema(testSchemaOne{})
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		_ = createTestTable(ctx, t, datasetID, tableID, s,
			cloudBQ.TableMetadataToUpdate{})
		md, err := GetTableMetadata(ctx, datasetID, tableID)
		want := &cloudBQ.TableMetadata{
			ID:     tableID,
			Schema: s,
		}
		if !(reflect.DeepEqual(md, want)) {
			t.Logf("Want: %v Have %v", want, md)
			t.Fail()
		}
	})
}

func TestCreateTable(t *testing.T) {
	ctx := createContext()
	datasetID := "test_dataset"
	tableID := "test_table_1"
	s, err := cloudBQ.InferSchema(testSchemaOne{})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Run("WithoutDatasetID", func(t *testing.T) {
		table := getTestTable(ctx, "", tableID, t)
		err := CreateTable(ctx, s, table)
		if err == nil {
			t.Log("Want: error")
			t.Log("Have: no error")
			t.Fail()
		}
		if tableWasCreated(ctx, datasetID, tableID) {
			t.Log("Want: table not created")
			t.Log("Have: table created")
			t.Fail()
		}
	})
	t.Run("WithoutTableID", func(t *testing.T) {
		table := getTestTable(ctx, datasetID, "", t)
		err := CreateTable(ctx, s, table)
		if err == nil {
			t.Log("Want: error")
			t.Log("Have: no error")
			t.Fail()
		}
		if tableWasCreated(ctx, datasetID, tableID) {
			t.Log("Want: table not created")
			t.Log("Have: table created")
			t.Fail()
		}
	})
	t.Run("Successful", func(t *testing.T) {
		table := getTestTable(ctx, datasetID, tableID, ctx, t)
		err := CreateTable(ctx, s, table)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		if !tableWasCreated(ctx, datasetID, tableID) {
			t.Log("Want: table created")
			t.Log("Have: table not created")
			t.Fail()
		}
	})
}

func TestUpdateTable(t *testing.T) {
	ctx := createContext()
	datasetID := "test_dataset"
	tableID := "test_table_1"
	s, err := cloudBQ.InferSchema(testSchemaOne{})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Run("WithoutDatasetID", func(t *testing.T) {
		table := getTestTable(ctx, "", tableID, t)
		toUpdate := cloudBQ.TableMetadataToUpdate{}
		_, err := UpdateTable(ctx, toUpdate, table)
		if err == nil {
			t.Fail()
		}
	})
	t.Run("WithoutTableID", func(t *testing.T) {
		table := getTestTable(ctx, datasetID, "", ctx, t)
		toUpdate := cloudBQ.TableMetadataToUpdate{}
		_, err := UpdateTable(ctx, toUpdate, table)
		if err == nil {
			t.Fail()
		}
	})
	t.Run("TableNotCreated", func(t *testing.T) {
		table := getTestTable(ctx, datasetID, tableID, ctx, t)
		toUpdate := cloudBQ.TableMetadataToUpdate{}
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
		toUpdate := cloudBQ.TableMetadataToUpdate{}
		md, err := UpdateTable(ctx, toUpdate, table)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		if md.Name != name || md.Description != desc ||
			!reflect.DeepEqual(md.Schema, s) {
			t.Logf("Want: %v %v %v", name, desc, s)
			t.Logf("Have: %v %v %v", md.Name, md.Description,
				md.Schema)
			t.Fail()
		}
	})
	t.Run("UpdateOneThing", func(t *testing.T) {
		name := "test name"
		desc := "test desc"
		table := createTestTableWithMetadata(ctx, name, desc, t,
			datasetID, tableID, s)
		newName := "new test name"
		toUpdate := cloudBQ.TableMetadataToUpdate{Name: newName}
		md, err := UpdateTable(ctx, toUpdate, table)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		if md.Name != newName {
			t.Logf("Want: %v", newName)
			t.Logf("Have: %v", md.Name)
			t.Fail()
		}
		if md.Description != desc ||
			!reflect.DeepEqual(md.Schema, s) {
			t.Logf("Want: %v %v", desc, s)
			t.Logf("Have: %v %v", md.Description, md.Schema)
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
		newS, err := cloudBQ.InferSchema(testSchemaTwo{})
		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}
		toUpdate := cloudBQ.TableMetadataToUpdate{
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
			t.Logf("Want: %v %v %v", newName, newDesc, newS)
			t.Logf("Have: %v %v %v", md.Name, md.Description,
				md.Schema)
			t.Fail()
		}
	})
}
