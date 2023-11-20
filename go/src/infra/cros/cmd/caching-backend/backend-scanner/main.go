// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// backend-scanner is a service to scan caching backends and update a k8s
// configmap which mounted by caching frontend (i.e. the load balancer) and
// backends. It also scales the caching downloader replicas according to the
// count of backend server.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"slices"
	"time"

	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sApplyCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	k8sApplyMetaV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sTypedAppsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	k8sTypedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var (
	port                 = flag.Int("port", 8000, "the port the scanner service to listen")
	scanIntervalDuration = flag.Duration("scan-interval", 30*time.Minute, "the interval in seconds to scan all backends")

	namespace = flag.String("namespace", "skylab", "the k8s namespace of caching service")
	appLabel  = flag.String("app-label", "caching-backend", "the app label of the backend pods to scan")
	cmName    = flag.String("configmap-name", "caching-config", "the configmap name to save the generated caching configuration")

	serviceIP   = flag.String("service-ip", "", "the caching service service IP")
	servicePort = flag.Uint("service-port", 8082, "the caching service service port")

	frontendInterface = flag.String("frontend-interface", "bond0", "The NIC name to bring up the frontend service IP")
	frontendLBAlgo    = flag.String("frontend-lb-algo", "rr", "The load balancing/schedulng algorithm of the frontend, see go/peep-fleet-caching-lb-algo for all options")

	cacheSizeInGB     = flag.Uint("cache-size", 750, "The size of cache space in GB.")
	l7Port            = flag.Uint("l7-port", 8083, "the port for caching service L7 balancing")
	otelTraceEndpoint = flag.String("otel-trace-endpoint", "", "The OTel collector endpoint for backend service tracing. e.g. http://127.0.0.1:4317")

	downloaderDeployment    = flag.String("backend-downloader-name", "downloader", "the deployment name of the downloader which download files from Google Cloud Storage")
	downloaderScalingFactor = flag.Int("backend-downloader-scaling-factor", 4, "the factor to scale the downloader deployment with backends, i.e. '<downlaoder_replica>=<factor> * <backend_count>'")
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

		backendConf: &nginxConf{
			CacheSizeInGB:     int(*cacheSizeInGB),
			Port:              int(*servicePort),
			L7Port:            int(*l7Port),
			OtelTraceEndpoint: *otelTraceEndpoint,
		},
		frontendConf: &keepalivedConf{
			ServiceIP:   *serviceIP,
			ServicePort: int(*servicePort),
			Interface:   *frontendInterface,
			LBAlgo:      *frontendLBAlgo,
		},
		downloaderScale: &downloaderScale{
			name:   *downloaderDeployment,
			factor: *downloaderScalingFactor,
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/update", s.updateHandler)
	// TODO(guocb): add smarter logic to /startup and /shutdown to update the
	// backend list more efficiently.
	mux.HandleFunc("/startup", s.updateHandler)
	mux.HandleFunc("/shutdown", s.updateHandler)
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

// backendScanner is for the backend scanner service, which scans the specified
// backends and generate/update specified configMap.
type backendScanner struct {
	namespace   string
	appLabel    string
	cmName      string
	serviceIP   string
	servicePort uint

	backendConf     *nginxConf
	frontendConf    *keepalivedConf
	downloaderScale *downloaderScale
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

	s.backendConf.L7Servers = ips
	nc, err := genConfig("nginx", nginxTemplate, s.backendConf)
	if err != nil {
		return fmt.Errorf("scan once: %w", err)
	}

	s.frontendConf.RealServers = ips
	kc, err := genConfig("keepalived", keepalivedTempalte, s.frontendConf)
	if err != nil {
		return fmt.Errorf("scan once: %w", err)
	}
	d := map[string]string{
		"service-ip":      s.serviceIP,
		"service-port":    fmt.Sprintf("%d", s.servicePort),
		"nginx.conf":      nc,
		"keepalived.conf": kc,
	}
	if err := updateConfigMap(clientset.CoreV1().ConfigMaps(s.namespace), s.cmName, s.namespace, d); err != nil {
		return fmt.Errorf("scan once: %w", err)
	}

	if err := scaleDeployment(clientset.AppsV1().Deployments(*namespace), s.downloaderScale.name, s.downloaderScale.getReplica(len(ips))); err != nil {
		return fmt.Errorf("update: %w", err)
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

// scaleDeployment scales the named deployment to the specified replica.
func scaleDeployment(c k8sTypedAppsV1.DeploymentInterface, name string, replica int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	scale, err := c.GetScale(ctx, name, k8sMetaV1.GetOptions{})
	if err != nil {
		return fmt.Errorf("scale deploy/%s: %w", name, err)
	}
	scale.Spec.Replicas = replica
	_, err = c.UpdateScale(ctx, name, scale, k8sMetaV1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("scale deploy/%s: %w", name, err)
	}
	return nil
}

// nginxConf is for rendering the Nginx configuration template.
type nginxConf struct {
	CacheSizeInGB     int      // the caching space size
	Port              int      // the backend port (for L4 balancing)
	L7Servers         []string // the L7 servers which cache files based on URI
	L7Port            int      // the port of L7 servers
	OtelTraceEndpoint string   // for OpenTelemetry instruments
}

// keepalivedConf is for rendering the Keepalived configuration template.
type keepalivedConf struct {
	ServiceIP   string   // the service IP (or VIP)
	ServicePort int      // the service port
	RealServers []string // the real server (i.e. backend) list
	Interface   string   // the interface Keepalived brings up the service IP
	LBAlgo      string   // the load balancing algorithm to use
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
		// Pod may has no IP assigned when it's in states like "Pending".
		if ip := p.Status.PodIP; ip != "" {
			ips = append(ips, ip)
		}
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

// genConfig generates the final template data.
func genConfig(templateName, configTmpl string, configData interface{}) (string, error) {
	var buf bytes.Buffer
	tmpl := template.Must(template.New(templateName).Parse(configTmpl))
	if err := tmpl.Execute(&buf, configData); err != nil {
		return "", fmt.Errorf("build config: %w", err)
	}
	return buf.String(), nil
}

type downloaderScale struct {
	name   string
	factor int
}

// getReplica returns the downloader replicas according to the given backend
// count. The repicas has a minimum.
func (d *downloaderScale) getReplica(backendCount int) int32 {
	var minimum int32 = 4
	if r := int32(backendCount * d.factor); r > minimum {
		return r
	}
	return minimum

}
