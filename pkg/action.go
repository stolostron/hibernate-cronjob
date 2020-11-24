// auther: github.com/jnpacker
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"os"

	hivev1 "github.com/openshift/hive/pkg/apis/hive/v1"
	hiveclient "github.com/openshift/hive/pkg/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//  patchStringValue specifies a json patch operation for a string.
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// Simple error function
func check_error(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func powerStateChange(clientSet *hiveclient.Clientset, clusterDeployment *hivev1.ClusterDeployment, powerState string) string {
	patch := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/powerState",
		Value: powerState,
	}}
	patchInBytes, _ := json.Marshal(patch)
	changedCD, err := clientSet.
		HiveV1().
		ClusterDeployments(clusterDeployment.Namespace).
		Patch(context.TODO(), clusterDeployment.Name, types.JSONPatchType, patchInBytes, v1.PatchOptions{})
	check_error(err)
	return string(changedCD.Spec.PowerState)
}

func main() {
	var kubeconfig *string

	// Determine what action to take Hibernating or Running
	var TAKE_ACTION = os.Getenv("TAKE_ACTION")
	if TAKE_ACTION == "" || (TAKE_ACTION != "Hibernating" && TAKE_ACTION != "Running") {
		panic("Environment variable TAKE_ACTION missing: " + TAKE_ACTION)
	}
	kubeconfig = flag.String("kubeconfig", "/home/jpacker/.kube/config", "")
	flag.Parse()

	var config *rest.Config
	var err error
	if _, err := os.Stat("/home/jpacker/.kube/config"); !os.IsNotExist(err) {
		fmt.Println("Connecting with local kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		fmt.Println("Connecting using In Cluster Config")
		config, err = rest.InClusterConfig()
	}
	check_error(err)

	// Create a client for the hiveV1 CustomResourceDefinitions
	hiveset, err := hiveclient.NewForConfig(config)
	check_error(err)

	// Grab all ClusterDeployments to change the state
	cds, err := hiveset.HiveV1().ClusterDeployments("").List(context.TODO(), v1.ListOptions{})
	for _, clusterDeployment := range cds.Items {
		if clusterDeployment.Labels["hibernate"] != "skip" && string(clusterDeployment.Spec.PowerState) != TAKE_ACTION {

			fmt.Print(TAKE_ACTION + ": " + clusterDeployment.Name)
			newPowerState := powerStateChange(hiveset, &clusterDeployment, TAKE_ACTION)

			// Check the new state and report a response
			if newPowerState == TAKE_ACTION {
				fmt.Println("  ✓")
			} else {
				fmt.Println("  X")
			}
		} else {
			fmt.Println("Skip    : " + clusterDeployment.Name + "  (currently " + string(clusterDeployment.Spec.PowerState) + ")")
		}
	}
}
