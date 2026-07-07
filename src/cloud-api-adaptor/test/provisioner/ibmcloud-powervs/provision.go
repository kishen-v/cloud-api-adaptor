//go:build ibmcloud_powervs

// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package ibmcloud_powervs

import (
	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
)

func init() {
	pv.NewProvisionerFunctions["ibmcloud-powervs"] = NewIBMCloudPowerVSProvisioner
	pv.NewInstallChartFunctions["ibmcloud-powervs"] = NewIBMCloudPowerVSInstallChart
}
