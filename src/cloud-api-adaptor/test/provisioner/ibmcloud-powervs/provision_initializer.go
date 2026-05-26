// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package ibmcloud_powervs

import (
	"os"
	"strings"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
	log "github.com/sirupsen/logrus"
)

type IBMPowerVSProperties struct {
	ApiKey                   string
	TestPodVMImage           string
	PowerVSZone              string
	PowerVSRegion            string
	PowerVSServiceInstanceID string
	PowerVSSSHKeyName        string
	PowerVSProcessorType     string
	PowerVSSystemType        string
	PowerVSMemory            string
	PowerVSProcessors        string
}

var IBMPowerVSProps = &IBMPowerVSProperties{}

func InitIBMCloudProperties(properties map[string]string) error {

	IBMPowerVSProps = &IBMPowerVSProperties{
		ApiKey:                   properties["IBMCLOUD_API_KEY"],
		PowerVSRegion:            properties["POWERVS_REGION"],
		PowerVSZone:              properties["POWERVS_ZONE"],
		PowerVSServiceInstanceID: properties["POWERVS_SERVICE_INSTANCE_ID"],
		PowerVSSSHKeyName:        properties["POWERVS_SSH_KEY_NAME"],
		PowerVSMemory:            properties["POWERVS_MEMORY"],
		PowerVSProcessors:        properties["POWERVS_PROCESSORS"],
	}

	needProvisionStr := os.Getenv("TEST_PROVISION")
	if strings.EqualFold(needProvisionStr, "yes") || strings.EqualFold(needProvisionStr, "true") || pv.Action == "uploadimage" {
		if len(IBMPowerVSProps.ApiKey) <= 0 {
			log.Warn("[warning] IBMCLOUD_API_KEY was not set. Depending on environment variable.")
		}
	}
	return nil
}
