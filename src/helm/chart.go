// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package helm contains operations for working with helm charts.
package helm

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/zarf-dev/zarf/src/config"
	"github.com/zarf-dev/zarf/src/pkg/message"
)

// TemplateChart generates a helm template from a given chart.
func (h *Helm) TemplateChart(ctx context.Context) (manifest string, chartValues chartutil.Values, err error) {
	spinner := message.NewProgressSpinner("Templating helm chart %s", h.chart.Name)
	defer spinner.Stop()

	err = h.createActionConfig(h.chart.Namespace, spinner)

	// Setup K8s connection.
	if err != nil {
		return "", nil, fmt.Errorf("unable to initialize the K8s client: %w", err)
	}

	// Bind the helm action.
	client := action.NewInstall(h.actionConfig)

	client.DryRun = true
	client.Replace = true // Skip the name check.
	client.ClientOnly = true
	client.IncludeCRDs = true
	// TODO: Further research this with regular/OCI charts
	client.Verify = false
	client.InsecureSkipTLSverify = config.CommonOptions.InsecureSkipTLSVerify
	if h.kubeVersion != "" {
		parsedKubeVersion, err := chartutil.ParseKubeVersion(h.kubeVersion)
		if err != nil {
			return "", nil, fmt.Errorf("invalid kube version %s: %w", h.kubeVersion, err)
		}
		client.KubeVersion = parsedKubeVersion
	}
	client.ReleaseName = h.chart.ReleaseName

	// If no release name is specified, use the chart name.
	if client.ReleaseName == "" {
		client.ReleaseName = h.chart.Name
	}

	// Namespace must be specified.
	client.Namespace = h.chart.Namespace

	loadedChart, chartValues, err := h.loadChartData()
	if err != nil {
		return "", nil, fmt.Errorf("unable to load chart data: %w", err)
	}

	client.PostRenderer, err = h.newRenderer(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("unable to create helm renderer: %w", err)
	}

	// Perform the loadedChart installation.
	templatedChart, err := client.Run(loadedChart, chartValues)
	if err != nil {
		return "", nil, fmt.Errorf("error generating helm chart template: %w", err)
	}

	manifest = templatedChart.Manifest

	for _, hook := range templatedChart.Hooks {
		manifest += fmt.Sprintf("\n---\n%s", hook.Manifest)
	}

	spinner.Success()

	return manifest, chartValues, nil
}

func (h *Helm) loadChartData() (*chart.Chart, chartutil.Values, error) {
	var (
		loadedChart *chart.Chart
		chartValues chartutil.Values
		err         error
	)

	if h.chartOverride == nil {
		// If there is no override, get the chart and values info.
		loadedChart, err = h.loadChartFromTarball()
		if err != nil {
			return nil, nil, fmt.Errorf("unable to load chart tarball: %w", err)
		}

		chartValues, err = h.parseChartValues()
		if err != nil {
			return loadedChart, nil, fmt.Errorf("unable to parse chart values: %w", err)
		}
	} else {
		// Otherwise, use the overrides instead.
		loadedChart = h.chartOverride
		chartValues = h.valuesOverrides
	}

	return loadedChart, chartValues, nil
}
