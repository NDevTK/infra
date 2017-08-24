// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package eventupload is a library for streaming events to BigQuery.
package eventupload

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"infra/libs/bqschema/tabledef"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	"golang.org/x/net/context"
)

var id idGenerator

type idGenerator struct {
	counter int64
	prefix  string
}

// eventUploader is an interface for types which implement a Put method. It
// exists for the purpose of mocking Uploader in tests.
type eventUploader interface {
	Put(src interface{}) error
}

// Uploader contains the necessary data for streaming data to BigQuery.
type Uploader struct {
	ctx context.Context
	// Uploader is bound to a specific table. DatasetID and Table ID are
	// provided for reference.
	DatasetID string
	TableID   string
	u         *bigquery.Uploader
	s         bigquery.Schema
	timeout   time.Duration
	// uploads is the Counter metric described by UploadMetricName,
	// specified in UploaderConfig. It contains a field "status" set to
	// either "success" or "failure."
	uploads metric.Counter
}

// UploaderConfig holds the configuration for an uploader.
type UploaderConfig struct {
	// SkipInvalid is an option for bigquery.Uploader.
	SkipInvalid bool
	// IgnoreUnknown is an option for bigquery.Uploader.
	IgnoreUnknown bool
	// Timeout is used by Put to determine how long it should attempt to
	// upload rows to BigQuery. Put automatically retries on transient
	// errors, so a timeout is necessary to avoid indefinite retries in the
	// event the error is not actually transient. If no value (0s) is
	// specified, NewUploader() will set a default.
	Timeout time.Duration
	// UploadsMetricName is a string used to create a tsmon Counter metric
	// for event upload attempts via Put, e.g.
	// "/chrome/infra/commit_queue/events/count". Set UploadMetricName
	// before the first call to Stage. If left unset, no metric will be
	// created.
	UploadMetricName string
}

func init() {
	var prefix string
	h, err := os.Hostname()
	if err != nil {
		log.Printf(fmt.Sprintf("eventupload: ERROR: os.Hostname() returns %s", err))
	} else {
		prefix = fmt.Sprintf("%s:%d:%d", h, os.Getpid(), time.Now().UnixNano())
	}
	id = idGenerator{prefix: prefix}
}

// NewUploader constructs a new Uploader struct.
//
// DatasetID and TableID are provided to the BigQuery client to gain access to
// a particular table.
func NewUploader(ctx context.Context, c *bigquery.Client, td *tabledef.TableDef, cfg UploaderConfig) (*Uploader, error) {
	if id.prefix == "" {
		return nil, errors.New("error initializing prefix for insertID")
	}

	u := c.Dataset(td.GetDataset().ID()).Table(td.TableId).Uploader()
	u.SkipInvalidRows = cfg.SkipInvalid
	u.IgnoreUnknownValues = cfg.IgnoreUnknown

	timeout := cfg.Timeout
	if timeout == time.Duration(0) {
		timeout = time.Minute
	}

	var uploadCounter metric.Counter
	if cfg.UploadMetricName != "" {
		name := cfg.UploadMetricName
		desc := "Upload attempts; status is 'success' or 'failure'"
		field := field.String("status")
		uploadCounter = metric.NewCounterIn(ctx, name, desc, nil, field)
	}

	s := tabledef.BQSchema(td.Fields)

	return &Uploader{
		ctx,
		td.GetDataset().ID(),
		td.TableId,
		u,
		s,
		timeout,
		uploadCounter,
	}, nil
}

func (u *Uploader) updateUploads(count int64, status string) {
	if u.uploads == nil || count == 0 {
		return
	}
	err := u.uploads.Add(u.ctx, count, status)
	if err != nil {
		log.Printf("eventupload: metric.Counter.Add: %v", err)
	}
}

// Put uploads one or more rows to the BigQuery service. src is expected to
// be a struct (or a pointer to a struct) matching the schema in Uploader, or a
// slice containing such values. Put takes care of adding InsertIDs, used by
// BigQuery to deduplicate rows.
//
// If any rows do now match one of the expected types, Put will not attempt to
// upload any rows and returns an InvalidTypeError.
//
// Put returns a PutMultiError if one or more rows failed to be uploaded.
// The PutMultiError contains a RowInsertionError for each failed row.
//
// Put will retry on temporary errors. If the error persists, the call will
// run indefinitely. Because of this, Put adds a timeout to the Context.
//
// See bigquery documentation and source code for detailed information on how
// struct values are mapped to rows.
func (u Uploader) Put(src interface{}) error {
	ctx, cancel := context.WithTimeout(u.ctx, u.timeout)
	defer cancel()
	var failed int64
	sss, err := prepareSrc(u.s, src)
	if err != nil {
		return err
	}
	err = u.u.Put(ctx, sss)
	if err != nil {
		log.Printf("eventupload: Uploader.Put: %v", err)
		if merr, ok := err.(bigquery.PutMultiError); ok {
			failed = int64(len(merr))
		} else {
			failed = int64(len(sss))
		}
		u.updateUploads(failed, "failure")
	}
	succeeded := int64(len(sss)) - failed
	if succeeded < 0 {
		log.Printf("eventupload: Uploader.Put succeeded < 0: %v", succeeded)
	} else {
		u.updateUploads(succeeded, "success")
	}
	return err
}

func prepareSrc(s bigquery.Schema, src interface{}) ([]*bigquery.StructSaver, error) {

	validateSingleValue := func(v reflect.Value) error {
		switch v.Kind() {
		case reflect.Struct:
			return nil
		case reflect.Ptr:
			ptrK := v.Elem().Kind()
			if ptrK == reflect.Struct {
				return nil
			}
			return errors.Reason("pointer types must point to structs, not %s", ptrK).Err()
		default:
			return errors.Reason("struct or pointer-to-struct expected, got %v", v.Kind()).Err()
		}
	}

	var srcs []interface{}

	srcV := reflect.ValueOf(src)
	if srcV.Kind() == reflect.Slice {
		srcs = make([]interface{}, srcV.Len())
		for i := 0; i < srcV.Len(); i++ {
			itemV := srcV.Index(i)
			if err := validateSingleValue(itemV); err != nil {
				return nil, err
			}
			srcs[i] = itemV.Interface()
		}
	} else {
		if err := validateSingleValue(srcV); err != nil {
			return nil, err
		}
		srcs = []interface{}{src}
	}

	var prepared []*bigquery.StructSaver
	for _, src := range srcs {
		ss := &bigquery.StructSaver{
			Schema:   s,
			InsertID: id.generateInsertID(),
			Struct:   src,
		}

		prepared = append(prepared, ss)
	}
	return prepared, nil
}

func (id *idGenerator) generateInsertID() string {
	counter := atomic.AddInt64(&id.counter, 1)
	return fmt.Sprintf("%s:%d", id.prefix, counter)
}

// BatchUploader contains the necessary data for asynchronously sending batches
// of event row data to BigQuery.
type BatchUploader struct {
	// TickC is a channel used by BatchUploader to prompt upload(). Set
	// TickC before the first call to Stage. A <-chan time.Time with a
	// ticker can be constructed with time.NewTicker(time.Duration).C. If
	// left unset, the default upload interval is one minute.
	TickC <-chan time.Time
	// tick holds the default ticker
	tick *time.Ticker

	u     eventUploader
	stopc chan struct{}
	wg    sync.WaitGroup

	mu      sync.Mutex
	pending []interface{}
	started bool
}

func (bu *BatchUploader) start() {
	if bu.TickC == nil {
		bu.tick = time.NewTicker(time.Minute)
		bu.TickC = bu.tick.C
	}

	bu.wg.Add(1)
	go func() {
		defer bu.wg.Done()
		for {
			select {
			case <-bu.TickC:
				bu.upload()
			case <-bu.stopc:
				return
			}
		}
	}()
}

// NewBatchUploader constructs a new BatchUploader, which may optionally be
// further configured by setting its exported fields before the first call to
// Stage. Its Close method should be called when it is no longer needed.
//
// Uploader implements eventUploader.
func NewBatchUploader(u eventUploader) (*BatchUploader, error) {
	bu := &BatchUploader{
		u:     u,
		stopc: make(chan struct{}),
	}
	return bu, nil
}

// Stage stages one or more rows for sending to BigQuery. src is expected to
// be a struct matching the schema in Uploader, or a slice containing
// such structs. Stage returns immediately and batches of rows will be sent to
// BigQuery at regular intervals according to the configuration of TickC.
func (bu *BatchUploader) Stage(src interface{}) {
	bu.mu.Lock()
	defer bu.mu.Unlock()

	if !bu.started {
		bu.start()
		bu.started = true
	}

	switch reflect.ValueOf(src).Kind() {
	case reflect.Slice:
		v := reflect.ValueOf(src)
		for i := 0; i < v.Len(); i++ {
			bu.pending = append(bu.pending, v.Index(i).Interface())
		}
	case reflect.Struct:
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

	_ = bu.u.Put(pending)
}

// Close flushes any pending event rows and releases any resources held by the
// uploader. Close should be called when the logger is no longer needed.
func (bu *BatchUploader) Close() {
	close(bu.stopc)
	bu.wg.Wait()

	// Final upload.
	bu.upload()

	if bu.tick != nil {
		bu.tick.Stop()
	}
}
