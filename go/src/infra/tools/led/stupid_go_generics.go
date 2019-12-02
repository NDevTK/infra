// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	bbpb "go.chromium.org/luci/buildbucket/proto"
	api "go.chromium.org/luci/swarming/proto/api"
)

// This file contains workarounds for the fact that Go doesn't have any useful
// generics.

type stringPair interface {
	GetPair() (key, value string)

	WithNewVal(value string) stringPair
}

type stringPairList interface {
	Len() int
	AddNew(k, v string)
	AddItem(stringPair)
	Copy() stringPairList
	Clear()
	Iter(cb func(stringPair))
}

type bbReqDims []*bbpb.RequestedDimension

func (r *bbReqDims) Len() int { return len(*r) }
func (r *bbReqDims) Clear()   { *r = (*r)[:0] }
func (r *bbReqDims) Copy() stringPairList {
	ret := make(bbReqDims, len(*r))
	copy(ret, *r)
	return &ret
}
func (r *bbReqDims) AddItem(itm stringPair) {
	*r = append(*r, (*bbpb.RequestedDimension)(itm.(*bbReqDim)))
}
func (r *bbReqDims) AddNew(k, v string) {
	*r = append(*r, &bbpb.RequestedDimension{Key: k, Value: v})
}
func (r *bbReqDims) Iter(cb func(stringPair)) {
	for _, itm := range *r {
		cb((*bbReqDim)(itm))
	}
}

type bbReqDim bbpb.RequestedDimension

func (r *bbReqDim) GetPair() (key, value string) { return r.Key, r.Value }
func (r *bbReqDim) WithNewVal(value string) stringPair {
	ret := *r
	ret.Value = value
	return &ret
}

// for the purposes of this, this is manipulated as if it were
// []*api.StringPair; "Values" only ever has one entry.
type swarmDims []*api.StringListPair

func (s *swarmDims) Len() int { return len(*s) }
func (s *swarmDims) Clear()   { *s = (*s)[:0] }
func (s *swarmDims) Copy() stringPairList {
	ret := make(swarmDims, len(*s))
	copy(ret, *s)
	return &ret
}
func (s *swarmDims) AddItem(itm stringPair) {
	*s = append(*s, (*api.StringListPair)(itm.(*swarmDim)))
}
func (s *swarmDims) AddNew(k, v string) {
	*s = append(*s, &api.StringListPair{Key: k, Values: []string{v}})
}
func (s *swarmDims) Iter(cb func(stringPair)) {
	for _, itm := range *s {
		cb((*swarmDim)(itm))
	}
}

type swarmDim api.StringListPair

func (s *swarmDim) GetPair() (key, value string) { return s.Key, s.Values[0] }
func (s *swarmDim) WithNewVal(value string) stringPair {
	return &swarmDim{Key: s.Key, Values: []string{value}}
}

type swarmEnvs []*api.StringPair

func (s *swarmEnvs) Len() int { return len(*s) }
func (s *swarmEnvs) Clear()   { *s = (*s)[:0] }
func (s *swarmEnvs) Copy() stringPairList {
	ret := make(swarmEnvs, len(*s))
	copy(ret, *s)
	return &ret
}
func (s *swarmEnvs) AddItem(itm stringPair) {
	*s = append(*s, (*api.StringPair)(itm.(*swarmEnv)))
}
func (s *swarmEnvs) AddNew(k, v string) {
	*s = append(*s, &api.StringPair{Key: k, Value: v})
}
func (s *swarmEnvs) Iter(cb func(stringPair)) {
	for _, itm := range *s {
		cb((*swarmEnv)(itm))
	}
}

type swarmEnv api.StringPair

func (s *swarmEnv) GetPair() (key, value string) { return s.Key, s.Value }
func (s *swarmEnv) WithNewVal(value string) stringPair {
	return &swarmEnv{Key: s.Key, Value: value}
}

func updateStringPairList(pairs stringPairList, updates map[string]string) {
	if len(updates) == 0 {
		return
	}
	myUpdates := make(map[string]string, len(updates))
	for k, v := range updates {
		myUpdates[k] = v
	}

	oldList := pairs.Copy()
	pairs.Clear()

	oldList.Iter(func(itm stringPair) {
		k, _ := itm.GetPair()
		if newVal, ok := myUpdates[k]; ok {
			delete(myUpdates, k)
			if newVal != "" {
				pairs.AddItem(itm.WithNewVal(newVal))
			}
		} else {
			pairs.AddItem(itm)
		}
	})

	for k, v := range myUpdates {
		if v != "" {
			pairs.AddNew(k, v)
		}
	}
}
