package main

import (
	"net/http"
	"reflect"

	cloudBQ "cloud.google.com/go/bigquery"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
)

const projectID = "chrome-infra-events"
const bqClientKey = "bqClient"

// TODO
// Instrument with ts_mon to count report errors

// item represents a test schema.
// TODO: Delete once we are using real schema
type item struct {
	Name  string
	Count int
}

type tableDef struct {
	datasetID  string
	tableID    string
	toUpdate cloudBQ.TableMetadataToUpdate
}

func getTableDefs() []tableDef {
	// TODO: Replace test data with data from luci-config
	schema, err := cloudBQ.InferSchema(item{})
	if err != nil {
		glog.Fatalf(err.Error())
	}
	return []tableDef{
		{
			datasetID: "test_dataset",
			tableID:   "new_table_3",
			toUpdate: cloudBQ.TableMetadataToUpdate{
				Description: "This is a test table",
				Name:        "New Table",
				Schema:      schema,
			},
		},
	}
}

func errNotFound(e error) bool {
	err, ok := e.(*googleapi.Error)
	return ok && err.Code == http.StatusNotFound
}

func updateFromTableDef(def tableDef, ctx context.Context) {
	t, err := getTable(ctx, def.datasetID, def.tableID)
	if err != nil {
		// Report error
		glog.Errorln(err.Error())
		return
	}
	md, err := getTableMetadata(ctx, def.datasetID, def.tableID)
	if err != nil  {
		if errNotFound(err) {
			if err := createTable(ctx, def.toUpdate.Schema, t); err != nil {
				// Report error to ts_mon
				glog.Errorln(err)
			}
		} else {
			// Report error to ts_mon
			glog.Errorln(err)
		}
	} else {
		// I need to pull the relevant fields out of md
		if !reflect.DeepEqual(md, def.toUpdate) {
			md, err = updateTable(ctx, def.toUpdate, t)
			if err != nil {
				glog.Errorln(err)
			}
		}
	}
}

func main() {
	ctx := context.Background()

	client, err := cloudBQ.NewClient(ctx, projectID)
	if err != nil {
		// Report to ts_mon
		glog.Errorln(err)
	}

	ctx = context.WithValue(ctx, bqClientKey, &prodBQ{c: client})

	defs := getTableDefs()
	for _, def := range defs {
		updateFromTableDef(def, ctx)
	}
}
