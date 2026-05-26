// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package ibmcloud_powervs

import (
	"errors"
	"os"
	"strconv"
	"strings"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
	log "github.com/sirupsen/logrus"
)

type IBMPowerVSProperties struct {
	IBMCloudProvider         string
	ApiKey                   string
	Region                   string
	WorkerOS                 string
	TestPodVMImage           string
	Zone                     string
	PowerVSZone              string
	PowerVSServiceInstanceID string
	PowerVSNetworkID         string
	PowerVSImageName         string
	PowerVSSSHKeyName        string
	PowerVSProcessorType     string
	PowerVSSystemType        string
	PowerVSMemory            string
	PowerVSProcessors        string
	KubernetesBuildVersion   string
	KubernetesReleaseMarker  string
	RetryOnTFFailure         int
	SSHPrivateKeyPath        string

	WorkerCount int
}

var IBMPowerVSProps = &IBMPowerVSProperties{}

func InitIBMCloudProperties(properties map[string]string) error {

	IBMPowerVSProps = &IBMPowerVSProperties{
		IBMCloudProvider:         properties["IBMCLOUD_PROVIDER"],
		ApiKey:                   properties["IBMCLOUD_API_KEY"],
		Region:                   properties["REGION"],
		WorkerOS:                 properties["WORKER_OS"],
		Zone:                     properties["ZONE"],
		PowerVSZone:              properties["POWERVS_ZONE"],
		PowerVSServiceInstanceID: properties["POWERVS_SERVICE_INSTANCE_ID"],
		PowerVSNetworkID:         properties["POWERVS_NETWORK_ID"],
		PowerVSImageName:         properties["POWERVS_IMAGE_NAME"],
		PowerVSSSHKeyName:        properties["POWERVS_SSH_KEY_NAME"],
		PowerVSProcessorType:     properties["POWERVS_PROCESSOR_TYPE"],
		PowerVSSystemType:        properties["POWERVS_SYSTEM_TYPE"],
		PowerVSMemory:            properties["POWERVS_MEMORY"],
		PowerVSProcessors:        properties["POWERVS_PROCESSORS"],
		KubernetesBuildVersion:   properties["KUBERNETES_BUILD_VERSION"],
		KubernetesReleaseMarker:  properties["KUBERNETES_RELEASE_MARKER"],
		SSHPrivateKeyPath:        properties["SSH_PRIVATE_KEY_PATH"],
	}

	if len(IBMPowerVSProps.IBMCloudProvider) <= 0 {
		IBMPowerVSProps.IBMCloudProvider = "ibmcloud-powervs"
	}

	workerCountStr := properties["WORKERS_COUNT"]
	if len(workerCountStr) <= 0 {
		IBMPowerVSProps.WorkerCount = 1
	} else {
		count, err := strconv.Atoi(workerCountStr)
		if err != nil {
			IBMPowerVSProps.WorkerCount = 1
		} else {
			IBMPowerVSProps.WorkerCount = count
		}
	}

	retryOnTFFailureStr := properties["RETRY_ON_TF_FAILURE"]
	if len(retryOnTFFailureStr) > 0 {
		retryCount, err := strconv.Atoi(retryOnTFFailureStr)
		if err == nil {
			IBMPowerVSProps.RetryOnTFFailure = retryCount
		}
	}

	if len(IBMPowerVSProps.Zone) <= 0 {
		if len(IBMPowerVSProps.PowerVSZone) > 0 {
			IBMPowerVSProps.Zone = IBMPowerVSProps.PowerVSZone
		} else {
			log.Info("[warning] ZONE was not set.")
		}
	}

	needProvisionStr := os.Getenv("TEST_PROVISION")
	if strings.EqualFold(needProvisionStr, "yes") || strings.EqualFold(needProvisionStr, "true") || pv.Action == "uploadimage" {
		if len(IBMPowerVSProps.ApiKey) <= 0 {
			return errors.New("APIKEY is required for provisioning")
		}
		if len(IBMPowerVSProps.Region) <= 0 {
			return errors.New("REGION was not set.")
		}
	}

	podvmImage := os.Getenv("TEST_PODVM_IMAGE")
	if len(podvmImage) > 0 {
		return nil
	}
	return nil
}
