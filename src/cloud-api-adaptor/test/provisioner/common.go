// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type PatchLabel struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// KindClusterScriptPath returns the absolute path to the canonical
// kind_cluster.sh script located in test/provisioner/common/.
// Using runtime.Caller(0) anchors the path to this source file's location,
// so callers receive the correct path regardless of the working directory.
func KindClusterScriptPath() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed to determine source path")
	}
	// thisFile is .../test/provisioner/common.go
	// kind_cluster.sh lives in .../test/provisioner/common/kind_cluster.sh
	script := filepath.Join(filepath.Dir(thisFile), "common", "kind_cluster.sh")
	return script, nil
}

var Action string

// KindClusterProperties holds the properties needed to manage a local kind cluster.
type KindClusterProperties struct {
	ClusterName      string
	ContainerRuntime string
	KindConfigFile   string
	WorkerNodeName   string
}

// KindCluster manages a local kind cluster for e2e testing.
type KindCluster struct {
	properties KindClusterProperties
}

// NewKindCluster creates a KindCluster from a properties map, applying sensible defaults.
func NewKindCluster(properties map[string]string) (*KindCluster, error) {
	clusterName := properties["CLUSTER_NAME"]
	if clusterName == "" {
		clusterName = "peer-pods-e2e"
	}
	kindConfigFile := properties["KIND_CONFIG_FILE"]
	containerRuntime := properties["CONTAINER_RUNTIME"]
	if containerRuntime == "" {
		containerRuntime = "containerd"
	}
	workerNodeName := properties["WORKER_NODE_NAME"]
	if workerNodeName == "" {
		workerNodeName = fmt.Sprintf("%s-worker", clusterName)
	}

	return &KindCluster{
		properties: KindClusterProperties{
			ClusterName:      clusterName,
			ContainerRuntime: containerRuntime,
			KindConfigFile:   kindConfigFile,
			WorkerNodeName:   workerNodeName,
		},
	}, nil
}

func (k *KindCluster) CreateCluster(ctx context.Context, cfg *envconf.Config) error {
	if k.properties.KindConfigFile == "" {
		return fmt.Errorf("KIND_CONFIG_FILE must be set to create a kind cluster")
	}
	kindConfigPath, err := filepath.Abs(k.properties.KindConfigFile)
	if err != nil {
		return fmt.Errorf("error getting absolute path of kind config file: %w", err)
	}

	log.Infof("Using kind config from: %s", kindConfigPath)

	if err := k.runScript("create", kindConfigPath); err != nil {
		log.Errorf("Error creating kind cluster: %v", err)
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	cfg.WithKubeconfigFile(filepath.Join(home, ".kube/config"))

	if err := AddNodeRoleWorkerLabel(context.Background(), k.properties.ClusterName, cfg); err != nil {
		return fmt.Errorf("failed to label nodes: %w", err)
	}

	// Update containerd configuration to not discard unpacked layers
	log.Info("Configuring containerd on worker node to keep unpacked layers...")

	cmd := exec.Command("docker", "exec", k.properties.WorkerNodeName, "sed", "-i",
		"s/discard_unpacked_layers = true/discard_unpacked_layers = false/g",
		"/etc/containerd/config.toml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warnf("Failed to update containerd config: %v, output: %s", err, string(output))
	} else {
		log.Info("Updated containerd config to keep unpacked layers")

		// Restart containerd to apply the change
		cmd = exec.Command("docker", "exec", k.properties.WorkerNodeName, "systemctl", "restart", "containerd")
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Warnf("Failed to restart containerd: %v, output: %s", err, string(output))
		} else {
			log.Info("Restarted containerd, waiting for it to be ready...")
			time.Sleep(5 * time.Second)

			// Verify if containerd is running
			cmd = exec.Command("docker", "exec", k.properties.WorkerNodeName, "systemctl", "is-active", "containerd")
			output, err = cmd.CombinedOutput()
			status := strings.TrimSpace(string(output))
			if err != nil || status != "active" {
				log.Warnf("Containerd may not be running properly: status=%s, err=%v", status, err)
			} else {
				log.Info("Containerd is active and running")
			}
		}
	}
	return nil
}

func (k *KindCluster) DeleteCluster(ctx context.Context, cfg *envconf.Config) error {
	return k.runScript("delete", "")
}

func (k *KindCluster) runScript(action, kindConfigPath string) error {
	scriptPath, err := KindClusterScriptPath()
	if err != nil {
		return fmt.Errorf("failed to locate kind_cluster.sh: %w", err)
	}
	cmd := exec.Command("/bin/bash", scriptPath, action)
	cmd.Stdout = os.Stdout
	// TODO: better handle stderr. Messages getting out of order.
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	// Set CLUSTER_NAME and CONTAINER_RUNTIME. Unset KUBECONFIG so the default path is used.
	cmd.Env = append(cmd.Env,
		"CLUSTER_NAME="+k.properties.ClusterName,
		"KUBECONFIG=",
		"CONTAINER_RUNTIME="+k.properties.ContainerRuntime,
	)
	if kindConfigPath != "" {
		cmd.Env = append(cmd.Env, "KIND_CONFIG_FILE="+kindConfigPath)
	}
	if err := cmd.Run(); err != nil {
		log.Errorf("Error running kind_cluster.sh %s: %v", action, err)
		return err
	}
	return nil
}

// Adds the worker label to all workers nodes in a given cluster
func AddNodeRoleWorkerLabel(ctx context.Context, clusterName string, cfg *envconf.Config) error {
	fmt.Printf("Adding worker label to nodes belonging to: %s\n", clusterName)
	client, err := cfg.NewClient()
	if err != nil {
		return err
	}

	nodelist := &corev1.NodeList{}
	if err := client.Resources().List(ctx, nodelist); err != nil {
		return err
	}
	// Use full path to avoid overwriting other labels (see RFC 6902)
	payload := []PatchLabel{{
		Op: "add",
		// "/" must be written as ~1 (see RFC 6901)
		Path:  "/metadata/labels/node.kubernetes.io~1worker",
		Value: "",
	}}
	payloadBytes, _ := json.Marshal(payload)
	workerStr := clusterName + "-worker"
	for _, node := range nodelist.Items {
		if strings.Contains(node.Name, workerStr) {
			if err := client.Resources().Patch(ctx, &node, k8s.Patch{PatchType: types.JSONPatchType, Data: payloadBytes}); err != nil {
				return err
			}
		}

	}
	return nil
}
