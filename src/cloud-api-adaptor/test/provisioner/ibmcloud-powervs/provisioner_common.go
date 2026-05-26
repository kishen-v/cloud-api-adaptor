// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package ibmcloud_powervs // IBMCloudPowerVSProvisioner implements the CloudProvisioner interface for ibmcloud PowerVS.

import (
	"context"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/e2e-framework/pkg/envconf"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
)

// IBMCloudPowerVSProvisioner implements CloudProvisioner for IBM Cloud PowerVS.
// Kind cluster lifecycle is handled by KindCluster; VM lifecycle is managed
// externally (real PowerVS instances pre-created in the workspace).
type IBMCloudPowerVSProvisioner struct {
	kind *KindCluster

	IBMCloudPowerVSAPIKey    string
	PowerVSZone              string
	PowerVSServiceInstanceId string
	PowerVSImageID           string
	PowerVSNetworkID         string
	PowerVSSSHKeyName        string
	PowerVSSystemType        string
	PowerVSMemory            string
	PowerVSProcessorType     string
	PowerVSProcessors        string
}

func (p *IBMCloudPowerVSProvisioner) CreateCluster(ctx context.Context, cfg *envconf.Config) error {
	log.Info("IBMCloudPowerVS: provisioning local kind cluster for e2e tests")
	if err := p.kind.CreateCluster(ctx, cfg); err != nil {
		return err
	}
	log.Info("IBMCloudPowerVS: kind cluster ready")
	return nil
}

func (p *IBMCloudPowerVSProvisioner) DeleteCluster(ctx context.Context, cfg *envconf.Config) error {
	log.Info("IBMCloudPowerVS: deleting local kind cluster")
	if err := p.kind.DeleteCluster(ctx, cfg); err != nil {
		return err
	}
	log.Info("IBMCloudPowerVS: kind cluster deleted")
	return nil
}

func (p *IBMCloudPowerVSProvisioner) CreateVPC(ctx context.Context, cfg *envconf.Config) error {
	return nil
}

func (p *IBMCloudPowerVSProvisioner) DeleteVPC(ctx context.Context, cfg *envconf.Config) error {
	return nil
}

func (p *IBMCloudPowerVSProvisioner) GetProperties(ctx context.Context, cfg *envconf.Config) map[string]string {
	return map[string]string{
		"IBMCLOUD_API_KEY":            p.IBMCloudPowerVSAPIKey,
		"POWERVS_ZONE":                p.PowerVSZone,
		"POWERVS_SERVICE_INSTANCE_ID": p.PowerVSServiceInstanceId,
		"POWERVS_IMAGE_ID":            p.PowerVSImageID,
		"POWERVS_NETWORK_ID":          p.PowerVSNetworkID,
		"POWERVS_SSH_KEY_NAME":        p.PowerVSSSHKeyName,
		"POWERVS_SYSTEM_TYPE":         p.PowerVSSystemType,
		"POWERVS_MEMORY":              p.PowerVSMemory,
		"POWERVS_PROCESSOR_TYPE":      p.PowerVSProcessorType,
		"POWERVS_PROCESSORS":          p.PowerVSProcessors,
		"CLUSTER_NAME":                p.kind.props.ClusterName,
	}
}

func (p *IBMCloudPowerVSProvisioner) UploadPodvm(imagePath string, ctx context.Context, cfg *envconf.Config) error {
	return nil
}

func NewIBMCloudPowerVSProvisioner(properties map[string]string) (pv.CloudProvisioner, error) {
	if err := InitIBMCloudPowerVSProperties(properties); err != nil {
		return nil, err
	}

	kind, err := newKindCluster(properties)
	if err != nil {
		return nil, err
	}

	memory := properties["POWERVS_MEMORY"]
	if memory == "" {
		memory = "2"
	}

	return &IBMCloudPowerVSProvisioner{
		kind:                     kind,
		IBMCloudPowerVSAPIKey:    properties["IBMCLOUD_API_KEY"],
		PowerVSZone:              properties["POWERVS_ZONE"],
		PowerVSServiceInstanceId: properties["POWERVS_SERVICE_INSTANCE_ID"],
		PowerVSImageID:           properties["POWERVS_IMAGE_ID"],
		PowerVSNetworkID:         properties["POWERVS_NETWORK_ID"],
		PowerVSSSHKeyName:        properties["POWERVS_SSH_KEY_NAME"],
		PowerVSSystemType:        properties["POWERVS_SYSTEM_TYPE"],
		PowerVSMemory:            memory,
		PowerVSProcessorType:     properties["POWERVS_PROCESSOR_TYPE"],
		PowerVSProcessors:        properties["POWERVS_PROCESSORS"],
	}, nil
}

func InitIBMCloudPowerVSProperties(properties map[string]string) error {
	return InitIBMCloudProperties(properties)
}
