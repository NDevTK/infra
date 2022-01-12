// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// k8sApply applies a YAML content to the K8s cluster the program running on.
// The implementation is mostly inspired by
// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
func k8sApply(ctx context.Context, content string) error {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("apply to k8s: %s", err)
	}
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return fmt.Errorf("apply to k8s: %s", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("apply to k8s: %s", err)
	}
	obj := &unstructured.Unstructured{}
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := decUnstructured.Decode([]byte(content), nil, obj)
	if err != nil {
		return fmt.Errorf("apply to k8s: %s", err)
	}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("apply to k8s: %s", err)
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		dr = dyn.Resource(mapping.Resource)
	}

	before, after, err := putResource(ctx, dr, obj)
	if err != nil {
		return fmt.Errorf("apply to k8s: %s", err)
	}
	if err := logChanges(before, after); err != nil {
		return fmt.Errorf("apply to k8s: %s", err)
	}
	return nil
}

// putResource puts (i.e. create or patch) the target resource.
func putResource(ctx context.Context, dr dynamic.ResourceInterface, obj *unstructured.Unstructured) (before, after *unstructured.Unstructured, err error) {
	before, err = dr.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Printf("Failed to get the resource (not created yet?): %s", err)
		log.Printf("Will try to create the resource")
		after, err = dr.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("put resource: create the resource: %s", err)
		}
		return nil, after, nil
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, nil, fmt.Errorf("put resource: %s", err)
	}
	// Overwrite all changes made by others, otherwise there might be conflicts
	// which cannot be resolved automatically.
	force := true
	after, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data,
		metav1.PatchOptions{FieldManager: "k8s_app_roller", Force: &force},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("put resource: patch the resource: %s", err)
	}
	return before, after, nil
}

func logChanges(before, after *unstructured.Unstructured) error {
	filterFields(before)
	filterFields(after)

	js, err := json.Marshal(after)
	if err != nil {
		return fmt.Errorf("log changes %s", err)
	}
	log.Printf("Applied resource:\n%s", js)

	d := cmp.Diff(before, after)
	if d == "" {
		log.Printf("Nothing changed")
		return nil
	}
	log.Printf("Changes: (-before, +after)\n%s", d)
	// TODO(guocb) log the change to BQ.
	return nil
}

// filterFields filters out fields that constantly changing even there's no
// real changes.
func filterFields(obj *unstructured.Unstructured) {
	if obj == nil {
		return
	}
	obj.Object["metadata"].(map[string]interface{})["managedFields"] = nil
	obj.Object["metadata"].(map[string]interface{})["resourceVersion"] = nil
}
