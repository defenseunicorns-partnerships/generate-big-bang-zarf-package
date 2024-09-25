// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// Package helm contains operations for working with helm charts.
package helm

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/zarf-dev/zarf/src/api/v1alpha1"
	"github.com/zarf-dev/zarf/src/config"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
)

// Helm is a config object for working with helm charts.
type Helm struct {
	chart      v1alpha1.ZarfChart
	chartPath  string
	valuesPath string

	timeout time.Duration
	retries int

	kubeVersion string

	chartOverride   *chart.Chart
	valuesOverrides map[string]any

	settings     *cli.EnvSettings
	actionConfig *action.Configuration
}

// Modifier is a function that modifies the Helm config.
type Modifier func(*Helm)

// New returns a new Helm config struct.
func New(chart v1alpha1.ZarfChart, chartPath string, valuesPath string, mods ...Modifier) *Helm {
	h := &Helm{
		chart:      chart,
		chartPath:  chartPath,
		valuesPath: valuesPath,
		timeout:    config.ZarfDefaultTimeout,
	}

	for _, mod := range mods {
		mod(h)
	}

	return h
}

// WithKubeVersion sets the Kube version for templating the chart
func WithKubeVersion(kubeVersion string) Modifier {
	return func(h *Helm) {
		h.kubeVersion = kubeVersion
	}
}

// StandardName generates a predictable full path for a helm chart for Zarf.
func StandardName(destination string, chart v1alpha1.ZarfChart) string {
	return filepath.Join(destination, chart.Name+"-"+chart.Version)
}

// StandardValuesName generates a predictable full path for the values file for a helm chart for zarf
func StandardValuesName(destination string, chart v1alpha1.ZarfChart, idx int) string {
	return fmt.Sprintf("%s-%d", StandardName(destination, chart), idx)
}
