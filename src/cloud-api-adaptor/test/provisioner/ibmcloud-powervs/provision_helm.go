//go:build ibmcloud_powervs

// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package ibmcloud_powervs

import (
	"context"
	"path/filepath"
	"strings"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func NewIBMCloudPowerVSInstallChart(installDir, provider string) (pv.InstallChart, error) {
	chartPath := filepath.Join(installDir, "charts", "peerpods")
	namespace := pv.GetCAANamespace()
	releaseName := "peerpods"
	debug := false
	// Convert provider name from ibmcloud_powervs to ibmcloud-powervs for helm values
	providerName := strings.ReplaceAll(provider, "_", "-")

	helm, err := pv.NewHelm(chartPath, namespace, releaseName, providerName, debug)
	if err != nil {
		return nil, err
	}

	return &IBMCloudPowerVSInstallChart{
		Helm: helm,
	}, nil
}

type IBMCloudPowerVSInstallChart struct {
	Helm *pv.Helm
}

func (ic *IBMCloudPowerVSInstallChart) Install(ctx context.Context, cfg *envconf.Config) error {
	return ic.Helm.Install(ctx, cfg)
}

func (ic *IBMCloudPowerVSInstallChart) Uninstall(ctx context.Context, cfg *envconf.Config) error {
	return ic.Helm.Uninstall(ctx, cfg)
}

func (ic *IBMCloudPowerVSInstallChart) Configure(ctx context.Context, cfg *envconf.Config, properties map[string]string) error {
	// Set provider-specific values from properties
	for key, value := range properties {
		ic.Helm.OverrideProviderValues[key] = value
	}
	return nil
}
