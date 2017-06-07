// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package eventupload is a library for streaming events to BigQuery.
package eventupload

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
)

var mu = &sync.Mutex{}
var idCounter int

// Uploader contains the necessary data for streaming data to BigQuery.
type Uploader struct {
	datasetID string
	tableID   string
	u         *bigquery.Uploader
	s         bigquery.Schema
	processID string
}

// NewUploader constructs a new Uploader struct.
func NewUploader(ctx context.Context, datasetID, tableID string, skipInvalid, ignoreUnknown bool) (*Uploader, error) {
	c, err := bigquery.NewClient(ctx, "chrome-infra-events")
	if err != nil {
		return nil, err
	}
	t := c.Dataset(datasetID).Table(tableID)
	md, err := t.Metadata(ctx)
	if err != nil {
		return nil, err
	}
	u := t.Uploader()
	u.SkipInvalidRows = skipInvalid
	u.IgnoreUnknownValues = ignoreUnknown
	h, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	processID := fmt.Sprintf("%s:%d:%d", h, os.Getpid(), time.Now().Unix())
	return &Uploader{
		datasetID,
		tableID,
		u,
		md.Schema,
		processID,
	}, nil
}

// Put uploads one or more rows to the BigQuery service. src is expected to
// be a struct matching the schema in Uploader, or a slice containing
// such structs. Put takes care of adding InsertIDs, used by BigQuery to
// deduplicate rows.
//
// Put returns a PutMultiError if one or more rows failed to be uploaded.
// The PutMultiError contains a RowInsertionError for each failed row.
//
// Put will retry on temporary errors. If the error persists, the call will
// run indefinitely. Because of this, Put adds a timeout to the Context.
//
// See bigquery documentation and source code for detailed information on how
// struct values are mapped to rows.
func (u Uploader) Put(ctx context.Context, src interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()
	if err := u.u.Put(ctx, prepareSrc(u.s, src, u.processID)); err != nil {
		return err
	}
	return nil
}

func prepareSrc(s bigquery.Schema, src interface{}, processID string) []*bigquery.StructSaver {
	var srcs []interface{}
	if sl, ok := src.([]interface{}); ok {
		srcs = sl
	} else {
		srcs = []interface{}{src}
	}

	var prepared []*bigquery.StructSaver
	for _, src := range srcs {
		ss := &bigquery.StructSaver{
			Schema:   s,
			InsertID: generateInsertID(processID),
			Struct:   src,
		}

		prepared = append(prepared, ss)
	}
	return prepared
}

func generateInsertID(processID string) string {
	var insertID string
	mu.Lock()
	insertID = fmt.Sprintf("%s:%d", processID, idCounter)
	idCounter++
	mu.Unlock()
	return insertID
}

// BatchUploader contains the necessary data for asynchronously sending batches
// of event row data to BigQuery.
type BatchUploader struct {
	u       Uploader
	ctx     context.Context
	tickc   <-chan time.Time
	stopc   chan struct{}
	mu      sync.Mutex
	pending []interface{}
	wg      sync.WaitGroup
}

// NewBatchUploader constructs a new BatchUploader. Its Close method should be
// called when it is no longer needed.
//
// uploadTicker controls the frequency with which batches of event row data are
// uploaded to BigQuery. It can be constructed with time.NewTicker().
func NewBatchUploader(ctx context.Context, u Uploader, uploadTicker <-chan time.Time) *BatchUploader {
	bu := &BatchUploader{
		u:     u,
		ctx:   ctx,
		stopc: make(chan struct{}),
		tickc: uploadTicker,
	}

	bu.wg.Add(1)
	go func() {
		defer bu.wg.Done()
		for {
			select {
			case <-bu.tickc:
				bu.upload()
			case <-bu.stopc:
				return
			}
		}
	}()
	return bu
}

// Stage stages one or more rows for sending to BigQuery. src is expected to
// be a struct matching the schema in Uploader, or a slice containing
// such structs. Stage returns immediately and batches of rows will be sent to
// BigQuery at regular intervals according to the configuration of tickc.
func (bu *BatchUploader) Stage(src interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if sl, ok := src.([]interface{}); ok {
		bu.pending = append(bu.pending, sl...)
	} else {
		bu.pending = append(bu.pending, src)
	}
}

// upload streams a batch of event rows to BigQuery. Put takes care of retrying,
// so if it returns an error there is either an issue with the data it is trying
// to upload, or BigQuery itself is experiencing a failure. So, we don't retry.
func (bu *BatchUploader) upload() {
	bu.mu.Lock()
	pending := bu.pending
	bu.pending = nil
	bu.mu.Unlock()

	if len(pending) == 0 {
		return
	}

	if err := bu.u.Put(bu.ctx, pending); err != nil {
		log.Printf("eventupload: WARNING: error from Put: %s", err)
	}
}

// Close flushes any pending event rows and releases any resources held by the
// uploader. Close should be called when the logger is no longer needed.
func (bu *BatchUploader) Close() {
	close(bu.stopc)
	bu.wg.Wait()

	// Final upload.
	bu.upload()
}
