//go:build ibmcloud_powervs

// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package ibmcloud_powervs

import (
	"context"
	"fmt"
	"path/filepath"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func NewIBMCloudPowerVSInstallChart(installDir, provider string) (pv.InstallChart, error) {
	chartPath := filepath.Join(installDir, "charts", "peerpods")
	namespace := pv.GetCAANamespace()
	releaseName := "peerpods"
	debug := false

	// The providers file on disk is ibmcloudpowervs.yaml and the chart provider value is
	// also "ibmcloudpowervs", matching what CLOUD_PROVIDER is set to when running tests.
	helm, err := pv.NewHelm(chartPath, namespace, releaseName, "ibmcloudpowervs", debug)
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

func (ic *IBMCloudPowerVSInstallChart) GetHelm() *pv.Helm {
	return ic.Helm
}

// Install creates the SSH key secret required by CAA, then installs the Helm
// chart and waits for the worker node to receive the kata-runtime label.
func (ic *IBMCloudPowerVSInstallChart) Install(ctx context.Context, cfg *envconf.Config) error {
	return ic.Helm.Install(ctx, cfg)
}

func (ic *IBMCloudPowerVSInstallChart) Uninstall(ctx context.Context, cfg *envconf.Config) error {
	return ic.Helm.Uninstall(ctx, cfg)
}

func (ic *IBMCloudPowerVSInstallChart) Configure(ctx context.Context, cfg *envconf.Config, properties map[string]string) error {
	// Must match the provider/providerConfigs/providerSecrets key in ibmcloudpowervs.yaml.
	const chartProvider = "ibmcloud-powervs"

	// Set IBM Cloud API key as a provider secret.
	if key := properties["IBMCLOUD_API_KEY"]; key != "" {
		ic.Helm.OverrideValues[fmt.Sprintf("providerSecrets.%s.IBMCLOUD_API_KEY", chartProvider)] = key
	}

	// Set PowerVS-specific provider config values from properties.
	providerKeys := []string{
		"POWERVS_ZONE",
		"POWERVS_SERVICE_INSTANCE_ID",
		"POWERVS_IMAGE_ID",
		"POWERVS_NETWORK_ID",
		"POWERVS_SSH_KEY_NAME",
		"POWERVS_MEMORY",
		"POWERVS_PROCESSORS",
		"POWERVS_PROCESSOR_TYPE",
		"POWERVS_SYSTEM_TYPE",
	}
	for _, k := range providerKeys {
		if v := properties[k]; v != "" {
			ic.Helm.OverrideValues[fmt.Sprintf("providerConfigs.%s.%s", chartProvider, k)] = v
		}
	}

	return nil
}
