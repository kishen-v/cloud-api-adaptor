// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package ibmcloud_powervs // IBMCloudPowerVSProvisioner implements the CloudProvisioner interface for ibmcloud PowerVS.

import (
	"context"
	"strconv"

	"sigs.k8s.io/e2e-framework/pkg/envconf"

	byomprov "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner/byom"
	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
)

// IBMCloudPowerVSProvisioner implements CloudProvisioner for IBM Cloud PowerVS.
// Cluster lifecycle (kind) is delegated to an embedded ByomProvisioner; this
// type only carries the PowerVS-specific properties used by the Helm chart.
type IBMCloudPowerVSProvisioner struct {
	// Embedded BYOM provisioner handles kind cluster creation/deletion.
	*byomprov.ByomProvisioner

	IBMCloudPowerVSAPIKey    string
	PowerVSRegion            string
	PowerVSZone              string
	PowerVSServiceInstanceId string
	PowerVSImageID           string
	PowerVSNetworkID         string
	PowerVSNetworkName       string
	PowerVSSSHKeyName        string
	PowerVSSystemType        string
	PowerVSMemory            string
	PowerVSProcessorType     string
	PowerVSProcessors        string
}

func (p *IBMCloudPowerVSProvisioner) CreateCluster(ctx context.Context, cfg *envconf.Config) error {
	return p.ByomProvisioner.CreateCluster(ctx, cfg)
}

func (p *IBMCloudPowerVSProvisioner) DeleteCluster(ctx context.Context, cfg *envconf.Config) error {
	return p.ByomProvisioner.DeleteCluster(ctx, cfg)
}

func (p *IBMCloudPowerVSProvisioner) CreateVPC(ctx context.Context, cfg *envconf.Config) error {
	return nil
}

func (p *IBMCloudPowerVSProvisioner) DeleteVPC(ctx context.Context, cfg *envconf.Config) error {
	return nil
}

func (p *IBMCloudPowerVSProvisioner) GetProperties(ctx context.Context, cfg *envconf.Config) map[string]string {
	return map[string]string{
		"IBMCLOUD_API_KEY":             p.IBMCloudPowerVSAPIKey,
		"POWERVS_REGION":               p.PowerVSRegion,
		"POWERVS_ZONE":                 p.PowerVSZone,
		"POWERVS_SERVICE_INSTANCE_ID":  p.PowerVSServiceInstanceId,
		"POWERVS_IMAGE_ID":             p.PowerVSImageID,
		"POWERVS_NETWORK_ID":           p.PowerVSNetworkID,
		"POWERVS_NETWORK_NAME":         p.PowerVSNetworkName,
		"POWERVS_SSH_KEY_NAME":         p.PowerVSSSHKeyName,
		"POWERVS_SYSTEM_TYPE":          p.PowerVSSystemType,
		"POWERVS_MEMORY":               p.PowerVSMemory,
		"POWERVS_PROCESSOR_TYPE":       p.PowerVSProcessorType,
		"POWERVS_PROCESSORS":           p.PowerVSProcessors,
		// BYOM fields required by the Helm chart's SSH key secret and worker-node logic
		"SSH_SECRET_PRIV_KEY_PATH":     byomprov.ByomProps.SSHSecretPrivKeyPath,
		"SSH_SECRET_PUB_KEY_PATH":      byomprov.ByomProps.SSHSecretPubKeyPath,
		"SSH_USERNAME":                 byomprov.ByomProps.SSHUsername,
		"CLUSTER_NAME":                 byomprov.ByomProps.ClusterName,
		"WORKER_NODE_NAME":             byomprov.ByomProps.WorkerNodeName,
		"CONTAINER_RUNTIME":            byomprov.ByomProps.ContainerRuntime,
	}
}

func (p *IBMCloudPowerVSProvisioner) UploadPodvm(imagePath string, ctx context.Context, cfg *envconf.Config) error {
	return nil
}

func NewIBMCloudPowerVSProvisioner(properties map[string]string) (pv.CloudProvisioner, error) {
	if err := InitIBMCloudPowerVSProperties(properties); err != nil {
		return nil, err
	}

	// Build the embedded BYOM provisioner (handles kind cluster lifecycle).
	byomBase, err := byomprov.NewByomProvisioner(properties)
	if err != nil {
		return nil, err
	}

	memory := properties["POWERVS_MEMORY"]
	if memory == "" {
		memory = strconv.Itoa(IBMPowerVSProps.WorkerCount) // not used for memory – keep original default
		memory = "2"
	}

	return &IBMCloudPowerVSProvisioner{
		ByomProvisioner:          byomBase.(*byomprov.ByomProvisioner),
		IBMCloudPowerVSAPIKey:    properties["IBMCLOUD_API_KEY"],
		PowerVSRegion:            properties["POWERVS_REGION"],
		PowerVSZone:              properties["POWERVS_ZONE"],
		PowerVSServiceInstanceId: properties["POWERVS_SERVICE_INSTANCE_ID"],
		PowerVSImageID:           properties["POWERVS_IMAGE_ID"],
		PowerVSNetworkID:         properties["POWERVS_NETWORK_ID"],
		PowerVSNetworkName:       properties["POWERVS_NETWORK_NAME"],
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
