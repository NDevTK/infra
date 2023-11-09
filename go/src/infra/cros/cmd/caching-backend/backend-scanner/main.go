// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// backend-scanner is a service to scan caching backends and update a k8s
// configmap which mounted by caching frontend (i.e. the load balancer) and
// backends. It also scales the caching downloader replicas according to the
// count of backend server.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sApplyCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	k8sApplyMetaV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sTypedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var (
	port                 = flag.Int("port", 80, "the port the scanner service to listen")
	scanIntervalDuration = flag.Duration("scan-interval", 30*time.Minute, "the interval in seconds to scan all backends")

	namespace = flag.String("namespace", "skylab", "the k8s namespace of caching service")
	appLabel  = flag.String("app-label", "caching-backend", "the app label of the backend pods to scan")
	cmName    = flag.String("configmap-name", "caching-config", "the configmap name to save the generated caching configuration")

	serviceIP   = flag.String("service-ip", "", "the caching service service IP")
	servicePort = flag.Uint("service-port", 8082, "the caching service service port")
)

func main() {
	if err := innerMain(); err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Printf("Done")
}

func innerMain() error {
	flag.Parse()

	s := &backendScanner{
		namespace:   *namespace,
		appLabel:    *appLabel,
		cmName:      *cmName,
		serviceIP:   *serviceIP,
		servicePort: *servicePort,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/update", s.updateHandler)
	svr := http.Server{Addr: fmt.Sprintf(":%d", *port), Handler: mux}
	log.Printf("starting backend scanner")

	if err := s.scanOnce(); err != nil {
		return err
	}
	log.Printf("Initial configmap created")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.regularlyScan(ctx, *scanIntervalDuration)
	log.Printf("Regular scanning scheduled at every %s", *scanIntervalDuration)

	err := svr.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	return err
}

type backendScanner struct {
	namespace   string
	appLabel    string
	cmName      string
	serviceIP   string
	servicePort uint
}

func (s *backendScanner) updateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	if err := s.scanOnce(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Printf("Error writing the response body: %s", err)
		}
	}
}

// scanOnce scans the specified caching backends and run the defined actions.
func (s *backendScanner) scanOnce() error {
	clientset, err := getK8sClientSet()
	if err != nil {
		return fmt.Errorf("scan once: %w", err)
	}
	ips, err := getPodIPs(clientset.CoreV1().Pods(s.namespace), s.appLabel)
	if err != nil {
		return fmt.Errorf("scan once: %w", err)
	}
	log.Printf("Found pod ip (app=%s): %v", s.appLabel, ips)

	d := map[string]string{
		"service-ip":   s.serviceIP,
		"service-port": fmt.Sprintf("%d", s.servicePort),
	}
	if err := updateConfigMap(clientset.CoreV1().ConfigMaps(s.namespace), s.cmName, s.namespace, d); err != nil {
		return fmt.Errorf("scan once: %w", err)
	}
	return nil
}

// regularlyScan scans at the specified interval duration.
func (s *backendScanner) regularlyScan(ctx context.Context, d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("Regular scanning stopped")
			return
		case <-ticker.C:
			log.Printf("Starting a regular scan")
			if err := s.scanOnce(); err != nil {
				log.Printf("The regular scan failed: %s", err)
			}
		}
	}
}

// updateConfigMap updates the specified k8s configmap with the given data.
func updateConfigMap(c k8sTypedCoreV1.ConfigMapInterface, name, namespace string, data map[string]string) error {
	kind := "ConfigMap"
	version := "v1"
	cm := k8sApplyCoreV1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: k8sApplyMetaV1.TypeMetaApplyConfiguration{
			Kind:       &kind,
			APIVersion: &version,
		},
		ObjectMetaApplyConfiguration: &k8sApplyMetaV1.ObjectMetaApplyConfiguration{
			Name:      &name,
			Namespace: &namespace,
		},
		Data: data,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := c.Apply(ctx, &cm, k8sMetaV1.ApplyOptions{FieldManager: "caching-backend-scanner"}); err != nil {
		return fmt.Errorf("update configmap: %w", err)
	}
	return nil
}

// getPodIPs gets IP list of selected pods with the specified label.
func getPodIPs(p k8sTypedCoreV1.PodInterface, appLabel string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pods, err := p.List(ctx,
		k8sMetaV1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", appLabel),
		})
	if err != nil {
		log.Printf("List pods (app=%s): %s", appLabel, err)
		return nil, err
	}
	var ips []string
	for _, p := range pods.Items {
		ips = append(ips, p.Status.PodIP)
	}
	slices.Sort(ips) // to keep the list stable

	return ips, nil
}

func getK8sClientSet() (*kubernetes.Clientset, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("get K8s client set: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("get K8s client set: %w", err)
	}
	return clientset, nil
}
