package main

import (
	"reflect"
	"testing"

	cloudBQ "cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
)

// TestSchemaOne can be used with bigquery.InferSchema to create a test Schema.
type testSchemaOne struct {
	FieldOne string
	FieldTwo int
}

// TestSchemaTwo is like TestSchemaOne, but contains slighlty different field
// definitions.
type testSchemaTwo struct {
	FieldOne int
	FieldTwo int
}

type testCase struct {
	table  *cloudBQ.Table
	wantMD *cloudBQ.TableMetadata
	def    tableDef
	name   string
}

func createTestTable(datasetID string, tableID string,
		ctx context.Context, t *testing.T) *cloudBQ.Table {
	s, err := cloudBQ.InferSchema(testSchemaOne{})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	table := &cloudBQ.Table{
		DatasetID: datasetID,
		TableID:   tableID,
	}
	err = createTable(ctx, s, table)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	return table
}

func TestUpdateFromTableDef(t *testing.T) {
	client := newMockClient()
	ctx := context.Background()
	ctx = context.WithValue(ctx, bqClientKey, client)

	datasetID := "test_dataset"
	schemaOne, err := cloudBQ.InferSchema(testSchemaOne{})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	schemaTwo, err := cloudBQ.InferSchema(testSchemaTwo{})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	testCases := []testCase{
		{
			name:  "Test with different schema",
			table: createTestTable(datasetID, "1", ctx, t),
			wantMD: &cloudBQ.TableMetadata{
				ID:     "1",
				Schema: schemaTwo,
			},
			def: tableDef{
				datasetID: datasetID,
				tableID:   "1",
				toUpdate: cloudBQ.TableMetadataToUpdate{
					Schema: schemaTwo,
				},
			},
		},
		{
			name:  "Test noop",
			table: createTestTable(datasetID, "2", ctx, t),
			wantMD: &cloudBQ.TableMetadata{
				ID:     "2",
				Schema: schemaOne,
			},
			def: tableDef{
				datasetID: datasetID,
				tableID:   "2",
				toUpdate: cloudBQ.TableMetadataToUpdate{
					Schema: schemaOne,
				},
			},
		},
		{
			name:  "Test different name and schema",
			table: createTestTable(datasetID, "3", ctx, t),
			wantMD: &cloudBQ.TableMetadata{
				ID:     "3",
				Name:   "Test Table 3",
				Schema: schemaTwo,
			},
			def: tableDef{
				datasetID: datasetID,
				tableID:   "3",
				toUpdate: cloudBQ.TableMetadataToUpdate{
					Schema: schemaTwo,
					Name:   "Test Table 3",
				},
			},
		},
		{
			name:  "Test create table",
			table: &cloudBQ.Table{
				TableID: "4",
				DatasetID: datasetID,
			},
			wantMD: &cloudBQ.TableMetadata{
				ID:     "4",
				Schema: schemaTwo,
			},
			def: tableDef{
				datasetID: datasetID,
				tableID:   "4",
				toUpdate: cloudBQ.TableMetadataToUpdate{
					Schema: schemaTwo,
				},
			},
		},
	}

	for _, testCase := range testCases {
		updateFromTableDef(testCase.def, ctx)
		md, err := getTableMetadata(
			ctx, "test_dataset", testCase.table.TableID)
		if err != nil {
			t.Logf(err.Error())
			t.Fail()
		}
		if !reflect.DeepEqual(testCase.wantMD, md) {
			t.Logf("Test case failed: %v", testCase.name)
			t.Logf("Want: %v Have: %v", testCase.wantMD, md)
			t.Fail()
		}
	}
}
