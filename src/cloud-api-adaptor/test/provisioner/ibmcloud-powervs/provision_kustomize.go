//go:build ibmcloud_powervs

// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package ibmcloud_powervs

import (
	"context"
	"path/filepath"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func NewIBMCloudPowerVSInstallOverlay(installDir, provider string) (pv.InstallOverlay, error) {
	overlay, err := pv.NewKustomizeOverlay(filepath.Join(installDir, "overlays", provider))
	if err != nil {
		return nil, err
	}

	return &IBMCloudPowerVSInstallOverlay{
		Overlay: overlay,
	}, nil
}

type IBMCloudPowerVSInstallOverlay struct {
	Overlay *pv.KustomizeOverlay
}

func (iko *IBMCloudPowerVSInstallOverlay) Apply(ctx context.Context, cfg *envconf.Config) error {
	return iko.Overlay.Apply(ctx, cfg)
}

func (iko *IBMCloudPowerVSInstallOverlay) Delete(ctx context.Context, cfg *envconf.Config) error {
	return iko.Overlay.Delete(ctx, cfg)
}

func (iko *IBMCloudPowerVSInstallOverlay) Edit(ctx context.Context, cfg *envconf.Config, properties map[string]string) error {
	// Set ConfigMap values from properties
	for key, value := range properties {
		if err := iko.Overlay.SetKustomizeConfigMapGeneratorLiteral("peer-pods-cm", key, value); err != nil {
			return err
		}
	}
	return nil
}
