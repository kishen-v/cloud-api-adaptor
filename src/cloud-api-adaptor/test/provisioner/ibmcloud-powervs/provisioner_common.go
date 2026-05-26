package ibmcloud_powervs // IBMCloudPowerVSProvisioner implements the CloudProvisioner interface for ibmcloud PowerVS.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/e2e-framework/pkg/envconf"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
)

type IBMCloudPowerVSProvisioner struct {
	IBMCloudPowerVSAPIKey                                  string
	PowerVSRegion                                          string
	PowerVSClusterName                                     string
	PowerVSZone                                            string
	PowerVSServiceInstanceId                               string
	PowerVSImageID                                         string
	PowerVSNetworkName                                     string
	PowerVSImageName                                       string
	PowerVSSSHKeyName                                      string
	PowerVSSystemType                                      string
	PowerVSMemory, PowerVSProcessorType, PowerVSProcessors string
	SSHPrivateKeyPath                                      string
	WorkersCount                                           int
	RetryOnTFFailure                                       int

	KubernetesBuildVersion string
	KubeconfigPath         string
}

// ensureKubetest2TFInstalled checks if kubetest2-tf is installed and installs it if needed
func ensureKubetest2TFInstalled(ctx context.Context) error {
	log.Info("Checking if kubetest2-tf is installed...")

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		goPath = filepath.Join(homeDir, "go")
	}
	binDir := filepath.Join(goPath, "bin")

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Ensure binDir is in PATH for subsequent command executions
	if err := ensureBinDirInPath(binDir); err != nil {
		return err
	}

	if _, err := exec.LookPath("kubetest2-tf"); err != nil {
		log.Info("kubetest2-tf not found, downloading prebuilt binaries...")
		downloads := map[string]string{
			"https://provider-ibm-cloud-test-infra.s3.us.cloud-object-storage.appdomain.cloud/kubetest2-tf": filepath.Join(binDir, "kubetest2-tf"),
			"https://provider-ibm-cloud-test-infra.s3.us.cloud-object-storage.appdomain.cloud/terraform":    filepath.Join(binDir, "terraform"),
		}

		for url, destination := range downloads {
			if err := downloadExecutable(ctx, url, destination); err != nil {
				return err
			}
		}
	} else {
		log.Info("kubetest2-tf is already installed")
	}

	if err := ensureTerraformPluginsInstalled(ctx); err != nil {
		return err
	}

	return nil
}

// ensureBinDirInPath adds the bin directory to PATH if not already present
func ensureBinDirInPath(binDir string) error {
	currentPath := os.Getenv("PATH")

	// Check if binDir is already in PATH
	pathDirs := filepath.SplitList(currentPath)
	for _, dir := range pathDirs {
		if dir == binDir {
			log.Infof("Directory %s is already in PATH", binDir)
			return nil
		}
	}

	// Add binDir to PATH
	newPath := binDir + string(filepath.ListSeparator) + currentPath
	if err := os.Setenv("PATH", newPath); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}

	log.Infof("Added %s to PATH", binDir)
	return nil
}

func downloadExecutable(ctx context.Context, url, destination string) error {
	log.Infof("Downloading %s to %s", url, destination)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: HTTP %d", url, resp.StatusCode)
	}

	tempDestination := destination + ".tmp"
	out, err := os.Create(tempDestination)
	if err != nil {
		return fmt.Errorf("failed to create temporary file for %s: %w", destination, err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		_ = os.Remove(tempDestination)
		return fmt.Errorf("failed to save %s: %w", destination, err)
	}

	if err := out.Close(); err != nil {
		_ = os.Remove(tempDestination)
		return fmt.Errorf("failed to close %s: %w", destination, err)
	}

	if err := os.Chmod(tempDestination, 0755); err != nil {
		_ = os.Remove(tempDestination)
		return fmt.Errorf("failed to mark %s executable: %w", destination, err)
	}

	if err := os.Rename(tempDestination, destination); err != nil {
		_ = os.Remove(tempDestination)
		return fmt.Errorf("failed to install %s: %w", destination, err)
	}

	return nil
}

func ensureTerraformPluginsInstalled(ctx context.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pluginBaseDir := filepath.Join(homeDir, ".terraform.d", "plugins", "registry.terraform.io")
	downloads := map[string]string{
		"https://provider-ibm-cloud-test-infra.s3.us.cloud-object-storage.appdomain.cloud/plugins/IBM-Cloud/ibm/1.73.0/linux_ppc64le/terraform-provider-ibm":  filepath.Join(pluginBaseDir, "IBM-Cloud", "ibm", "1.73.0", "linux_ppc64le", "terraform-provider-ibm"),
		"https://provider-ibm-cloud-test-infra.s3.us.cloud-object-storage.appdomain.cloud/plugins/hashicorp/null/3.2.3/linux_ppc64le/terraform-provider-null": filepath.Join(pluginBaseDir, "hashicorp", "null", "3.2.3", "linux_ppc64le", "terraform-provider-null"),
	}

	for url, destination := range downloads {
		if _, err := os.Stat(destination); err == nil {
			log.Infof("Terraform plugin already installed at %s", destination)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			return fmt.Errorf("failed to create terraform plugin directory for %s: %w", destination, err)
		}

		if err := downloadExecutable(ctx, url, destination); err != nil {
			return err
		}
	}

	return nil
}

func (p *IBMCloudPowerVSProvisioner) CreateCluster(ctx context.Context, cfg *envconf.Config) error {
	log.Info("CreateCluster() - Using kubetest2-tf to provision PowerVS cluster on IBM Cloud")

	// Ensure kubetest2-tf is installed
	if err := ensureKubetest2TFInstalled(ctx); err != nil {
		return fmt.Errorf("failed to ensure kubetest2-tf is installed: %w", err)
	}

	// Generate cluster name with timestamp if not provided
	clusterName := p.PowerVSClusterName
	if clusterName == "" {
		clusterName = "cloud-api-adopter-e2e"
	}

	// Set default values
	if p.KubernetesBuildVersion == "" {
		p.KubernetesBuildVersion = "1.36.0"
	}

	if p.RetryOnTFFailure == 0 {
		p.RetryOnTFFailure = 3
	}
	if p.PowerVSMemory == "" {
		p.PowerVSMemory = "16"
	}
	// Detect API key from environment if not set
	if p.IBMCloudPowerVSAPIKey == "" {
		log.Info("Fetching the IBMCLOUD_API_KEY from the environment variable.")
		p.IBMCloudPowerVSAPIKey = os.Getenv("IBMCLOUD_API_KEY")
	}

	// Build kubetest2 command
	args := []string{
		"tf",
		"--powervs-image-name", p.PowerVSImageName,
		"--powervs-zone", p.PowerVSZone,
		"--powervs-region", p.PowerVSRegion,
		"--powervs-service-id", p.PowerVSServiceInstanceId,
		"--powervs-network-name", p.PowerVSNetworkName,
		"--powervs-api-key", p.IBMCloudPowerVSAPIKey,
		"--powervs-ssh-key", p.PowerVSSSHKeyName,
		"--build-version", p.KubernetesBuildVersion,
		"--release-marker", p.KubernetesBuildVersion,
		"--kubeconfig-path", p.KubeconfigPath,
		"--cluster-name", clusterName,
		"--ignore-cluster-dir",
		"--workers-count", strconv.Itoa(p.WorkersCount),
		"--extra-vars=kubelet_extra_args:\"--runtime-request-timeout=20m\"",
		"--up",
		"--auto-approve",
		"--retry-on-tf-failure", strconv.Itoa(p.RetryOnTFFailure),
		"--break-kubetest-on-upfail", "true",
		"--powervs-memory", p.PowerVSMemory,
		"--extra-vars=directory:release",
	}

	// Add SSH private key if provided
	if p.SSHPrivateKeyPath != "" {
		args = append(args, "--ssh-private-key", p.SSHPrivateKeyPath)
	}

	log.Infof("Running kubetest2-tf command to setup kubernetes cluster...")

	cmd := exec.CommandContext(ctx, "kubetest2-tf", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create cluster with kubetest2-tf: %w", err)
	}
	// cmd = exec.CommandContext(ctx, fmt.Sprintf("kubectl label nodes %s-master node.kubernetes.io/worker=", p.PowerVSClusterName), args...)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	log.Info("Cluster created successfully")
	p.PowerVSClusterName = clusterName
	p.addWorkerLabel(ctx)
	cfg.WithKubeconfigFile(p.KubeconfigPath)
	return nil
}

func (p *IBMCloudPowerVSProvisioner) addWorkerLabel(ctx context.Context) error {
	workerSelector := "!node-role.kubernetes.io/control-plane"
	getWorkersCmd := exec.CommandContext(ctx, "kubectl", "get", "nodes", "-l", workerSelector, "-o", "jsonpath={.items[*].metadata.name}")

	workerOutput, err := getWorkersCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to poll for dedicated worker nodes: %w", err)
	}

	trimmedWorkers := strings.TrimSpace(string(workerOutput))
	var targetNodes []string

	if trimmedWorkers != "" {
		targetNodes = strings.Fields(trimmedWorkers)
	} else {
		getAllNodesCmd := exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", "jsonpath={.items[*].metadata.name}")
		allOutput, err := getAllNodesCmd.Output()
		if err != nil {
			return fmt.Errorf("failed fallback query for all nodes: %w", err)
		}

		trimmedAll := strings.TrimSpace(string(allOutput))
		if trimmedAll == "" {
			return fmt.Errorf("no nodes found in the cluster to label")
		}
		targetNodes = strings.Fields(trimmedAll)
	}

	for _, node := range targetNodes {
		labelArgs := []string{"label", "nodes", node, "node.kubernetes.io/worker="}
		cmd := exec.CommandContext(ctx, "kubectl", labelArgs...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to apply worker label on node %s: %w (output: %s)", node, err, string(output))
		}
	}
	log.Info("Labelled nodes successfully")
	return nil
}

func (p *IBMCloudPowerVSProvisioner) DeleteCluster(ctx context.Context, cfg *envconf.Config) error {
	log.Info("DeleteCluster() - Using kubetest2-tf to delete PowerVS cluster")

	// Detect API key from environment if not set
	if p.IBMCloudPowerVSAPIKey == "" {
		p.IBMCloudPowerVSAPIKey = os.Getenv("IBMCLOUD_API_KEY")
	}

	// Build kubetest2 command for deletion
	args := []string{
		"--powervs-zone", p.PowerVSZone,
		"--powervs-region", p.PowerVSRegion,
		"--powervs-service-id", p.PowerVSServiceInstanceId,
		"--ignore-cluster-dir",
		"--powervs-api-key", p.IBMCloudPowerVSAPIKey,
		"--cluster-name", p.PowerVSClusterName,
		"--down",
		"--auto-approve",
		"--ignore-destroy-errors",
	}

	log.Infof("Running kubetest2-tf command: %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "kubetest2-tf", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		log.Warnf("Failed to delete cluster with kubetest2-tf (continuing anyway): %v", err)
	} else {
		log.Info("Cluster deleted successfully")
	}
	return nil
}

func (p *IBMCloudPowerVSProvisioner) CreateVPC(ctx context.Context, cfg *envconf.Config) error {
	log.Trace("CreateVPC()")
	return nil
}

func (p *IBMCloudPowerVSProvisioner) DeleteVPC(ctx context.Context, cfg *envconf.Config) error {
	log.Trace("DeleteVPC()")
	return nil
}

func (p *IBMCloudPowerVSProvisioner) GetProperties(ctx context.Context, cfg *envconf.Config) map[string]string {
	return map[string]string{}
}
func (p *IBMCloudPowerVSProvisioner) UploadPodvm(imagePath string, ctx context.Context, cfg *envconf.Config) error {
	log.Trace("Image is expected to be already uploaded to the PowerVS Workspace")
	return nil
}

func NewIBMCloudPowerVSProvisioner(properties map[string]string) (pv.CloudProvisioner, error) {
	if err := InitIBMCloudPowerVSProperties(properties); err != nil {
		return nil, err
	}

	// Parse workers count (default to 0 - no worker nodes needed)
	workersCount := 0
	if val := properties["WORKERS_COUNT"]; val != "" {
		if count, err := strconv.Atoi(val); err == nil {
			workersCount = count
		}
	}

	// Parse retry on TF failure
	retryOnTFFailure := 3
	if val := properties["RETRY_ON_TF_FAILURE"]; val != "" {
		if retry, err := strconv.Atoi(val); err == nil {
			retryOnTFFailure = retry
		}
	}

	return &IBMCloudPowerVSProvisioner{
		IBMCloudPowerVSAPIKey:    properties["IBMCLOUD_API_KEY"],
		PowerVSRegion:            properties["POWERVS_REGION"],
		PowerVSClusterName:       properties["CLUSTER_NAME"],
		PowerVSZone:              properties["POWERVS_ZONE"],
		PowerVSNetworkName:       properties["POWERVS_NETWORK_NAME"],
		PowerVSServiceInstanceId: properties["POWERVS_SERVICE_INSTANCE_ID"],
		KubernetesBuildVersion:   properties["KUBERNETES_BUILD_VERSION"],
		KubeconfigPath:           properties["KUBECONFIG_PATH"],
		PowerVSImageName:         properties["POWERVS_IMAGE_NAME"],
		PowerVSSSHKeyName:        properties["POWERVS_SSH_KEY_NAME"],
		PowerVSMemory:            properties["POWERVS_WORKER_MEMORY"],
		PowerVSProcessors:        properties["POWERVS_WORKER_PROCESSORS"],
		SSHPrivateKeyPath:        properties["SSH_PRIVATE_KEY_PATH"],
		WorkersCount:             workersCount,
		RetryOnTFFailure:         retryOnTFFailure,
	}, nil
}

func InitIBMCloudPowerVSProperties(properties map[string]string) error {
	return InitIBMCloudProperties(properties)
}
