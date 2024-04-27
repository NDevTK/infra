// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/maruel/subcommands"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/grpc/codes"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"

	"infra/cmdsupport/cmdlib"
	fleetcostpb "infra/cros/fleetcost/api/models"
	"infra/cros/fleetcost/internal/fleetcosterror"
)

// ToIndicatorType converts a string to an indicator.
func ToIndicatorType(x string) (fleetcostpb.IndicatorType, error) {
	if x == "" {
		return fleetcostpb.IndicatorType_INDICATOR_TYPE_UNKNOWN, nil
	}
	out, err := lookupValue(fleetcostpb.IndicatorType_value, x, "INDICATOR_TYPE")
	return fleetcostpb.IndicatorType(out), err
}

const billion = 1000 * 1000 * 1000

// ToUSD converts a string on the command line to US dollars.
//
// Right now, it works by parsing a float. I'm not sure this is great, we
// really should be parsing the number as a big decimal without going through a
// float, but the resulting code for doing that, right now, ends up being more
// complex than is justified.
//
// I'm writing this comment partially to express my frustration at not knowing
// a simpler way to parse an arbitrary-precision big decimal from the command
// line (none of the stuff in math/big is an exact match) and partially to
// exhort my future self or other readers to replace this function with something
// better.
func ToUSD(x string) (*money.Money, error) {
	var val float64
	if _, err := fmt.Sscanf(x, "%f", &val); err != nil {
		return nil, fmt.Errorf("invalid number %q", x)
	}
	units := int64(val)
	// Extract the fractional part multiply by 1000 and round to the nearest integer.
	fPart := math.Round(1000 * (val - float64(units)))
	nanos := int32((billion / 1000) * fPart)
	return &money.Money{
		CurrencyCode: "USD",
		Units:        units,
		Nanos:        nanos,
	}, nil
}

func MoneyToFloat(v *money.Money) float64 {
	return float64(v.GetUnits()) + float64(v.GetNanos())/billion
}

// ToCostCadence converts a string to a cost cadence.
func ToCostCadence(x string) (fleetcostpb.CostCadence, error) {
	out, err := lookupValue(fleetcostpb.CostCadence_value, x, "COST_CADENCE")
	return fleetcostpb.CostCadence(out), err
}

// ToLocation converts a string to a location.
func ToLocation(x string) (fleetcostpb.Location, error) {
	if x == "" {
		return fleetcostpb.Location_LOCATION_UNKNOWN, nil
	}
	out, err := lookupValue(fleetcostpb.Location_value, x, "LOCATION")
	return fleetcostpb.Location(out), err
}

// Function lookValue looks up a string key in a proto map.
//
// First it uppercases the string, and adds a prefix if necessary, due to the verbosity of the proto naming conventions.
//
// The error that this function returns is intended to be multi-line and human readable.
func lookupValue(m map[string]int32, key string, prefix string) (int32, error) {
	key = strings.ToUpper(key)
	prefix = strings.ToUpper(prefix)
	candidates := []string{
		key,
		fmt.Sprintf("%s_%s", prefix, key),
	}
	for _, candidate := range candidates {
		if res, ok := m[candidate]; ok {
			return res, nil
		}
	}
	return 0, errors.New(strings.Join(lookupValueErrorMessage(m), "\n"))
}

// Function lookupValueErrorMessage creates a help message from a proto map.
func lookupValueErrorMessage(m map[string]int32) []string {
	var out []string
	out = append(out, "Choose a candidate from the following values (with or without prefix):")
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out = append(out, fmt.Sprintf("- %s", k))
	}
	return out
}

// RunPerhapsInTransaction runs a datastore command perhaps in a transaction.
func RunPerhapsInTransaction(ctx context.Context, callback func(context.Context) error, options *datastore.TransactionOptions) error {
	if datastore.CurrentTransaction(ctx) == nil {
		return datastore.RunInTransaction(ctx, callback, options)
	}
	return callback(ctx)
}

// ErrItemExists applies when we try to insert an item that already exists.
var ErrItemExists = errors.New("item already exists, cannot replace")

// InsertOneWithoutReplacement inserts an item without replacement.
//
// We insist on having a PropertyLoadSaver+MetaGetterSetter (rather than taking an any) because this function only inserts one thing without replacement.
// (It's not clear what the semantics should be if you want to replace multiple things without replacement).
func InsertOneWithoutReplacement(ctx context.Context, entity interface {
	datastore.PropertyLoadSaver
	datastore.MetaGetterSetter
}, options *datastore.TransactionOptions) error {
	return RunPerhapsInTransaction(ctx, func(ctx context.Context) error {
		existsResult, err := datastore.Exists(ctx, entity)
		if err != nil {
			return err
		}
		if existsResult.Any() {
			return fleetcosterror.WithDefaultCode(codes.AlreadyExists, ErrItemExists)
		}
		return datastore.Put(ctx, entity)
	}, options)
}

// DeleteOneIfExists deletes an entity if it exists.
func DeleteOneIfExists(ctx context.Context, entity interface {
	datastore.PropertyLoadSaver
	datastore.MetaGetterSetter
}, options *datastore.TransactionOptions) error {
	return RunPerhapsInTransaction(ctx, func(ctx context.Context) error {
		existsResult, err := datastore.Exists(ctx, entity)
		if err != nil {
			return err
		}
		if !existsResult.Any() {
			return fleetcosterror.WithDefaultCode(codes.NotFound, datastore.ErrNoSuchEntity)
		}
		return datastore.Delete(ctx, entity)
	}, options)
}

// PrintMultiError prints a multierror, unwrapping as necessary.
func PrintMultiError(a subcommands.Application, err error) {
	var merr errors.MultiError
	if errors.As(err, &merr) {
		for _, e := range merr {
			cmdlib.PrintError(a, e)
		}
	} else {
		cmdlib.PrintError(a, err)
	}
}

// MaybeErrorf logs an error if it's non-nil and then returns it.
func MaybeErrorf(ctx context.Context, e error) error {
	if e != nil {
		logging.Errorf(ctx, "%s\n", e)
	}
	return e
}

// ErrorStringContains is a helper function that returns true if and only if the error in question contains
// the substring msg.
func ErrorStringContains(e error, msg string) bool {
	if e == nil {
		return msg == ""
	}
	return strings.Contains(e.Error(), msg)
}
