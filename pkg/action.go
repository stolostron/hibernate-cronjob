// author: github.com/jnpacker
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	hivev1 "github.com/openshift/hive/pkg/apis/hive/v1"
	hiveclient "github.com/openshift/hive/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const HibernateSA = true
const ClusterInstallerSA = false

//  patchStringValue specifies a json patch operation for a string.
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// Simple error function
func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func powerStatePatch(clientSet *hiveclient.Clientset, clusterDeployment *hivev1.ClusterDeployment, powerState string) string {
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
	checkError(err)
	return string(changedCD.Spec.PowerState)
}

func powerStateUpdate(clientSet *hiveclient.Clientset, clusterDeployment *hivev1.ClusterDeployment, powerState string) string {
	clusterDeployment.Spec.PowerState = hivev1.ClusterPowerState(powerState)
	changedCD, err := clientSet.
		HiveV1().
		ClusterDeployments(clusterDeployment.Namespace).
		Update(context.TODO(), clusterDeployment, v1.UpdateOptions{})

	checkError(err)
	return string(changedCD.Spec.PowerState)
}

// Used to create events for Cluster hibernation actions
//objName, namespaceName, objKind, eventName, message, reason, eType, api_core
func fireEvent(clientSet *kubernetes.Clientset, objRef *hivev1.ClusterDeployment, eventName string, message string, reason string, eType string) {
	event, err := clientSet.EventsV1().Events(objRef.Namespace).Get(context.TODO(), eventName, v1.GetOptions{})
	if event != nil && event.Series == nil {
		event.Series = new(eventv1.EventSeries)
		event.Series.Count = 1
		event.Series.LastObservedTime = v1.NowMicro()
	}
	if err != nil {
		fmt.Println("  |-> Event not found")
		event = new(eventv1.Event)
		//event.TypeMeta.Kind = "Event"
		//event.TypeMeta.APIVersion = "v1"
		event.ObjectMeta.Name = eventName
		event.ObjectMeta.Namespace = objRef.Namespace
		event.EventTime = v1.NowMicro()
		event.Action = "hibernating"
		event.ReportingController = "hibernate-cronjob"
		event.ReportingInstance = objRef.Namespace + "/" + objRef.Name
	}
	if event.Series != nil {
		event.Series.Count = event.Series.Count + 1
		event.Series.LastObservedTime = v1.NowMicro()
	}
	event.Type = eType
	event.Reason = reason
	event.Note = message
	event.Regarding = corev1.ObjectReference{
		Kind:      objRef.Kind,
		Namespace: objRef.Namespace,
		Name:      objRef.Name,
	}
	if err != nil {
		_, err := clientSet.EventsV1().Events(objRef.Namespace).Create(context.TODO(), event, v1.CreateOptions{})
		fmt.Println("  \\-> Create a new event " + eventName + " for cluster " + objRef.Namespace + "/" + objRef.Name)
		checkError(err)
	} else {
		_, err := clientSet.EventsV1().Events(objRef.Namespace).Update(context.TODO(), event, v1.UpdateOptions{})
		fmt.Println("  \\-> Update existing event "+eventName+", event count", event.Series.Count)
		checkError(err)
	}
}

func main() {
	var kubeconfig *string

	// Determine what action to take Hibernating or Running
	var TakeAction = strings.ToLower(os.Getenv("TAKE_ACTION"))
	var OptIn = os.Getenv("OPT_IN")
	if TakeAction == "" || (TakeAction != "hibernating" && TakeAction != "running") {
		panic("Environment variable TAKE_ACTION missing: " + TakeAction)
	}
	TakeAction = strings.ToUpper(TakeAction[0:1]) + TakeAction[1:]

	homePath := os.Getenv("HOME") // Used to look for .kube/config
	kubeconfig = flag.String("kubeconfig", homePath+"/.kube/config", "")
	flag.Parse()

	var config *rest.Config
	var err error
	if _, err := os.Stat(homePath + "/.kube/config"); !os.IsNotExist(err) {
		fmt.Println("Connecting with local kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		fmt.Println("Connecting using In Cluster Config")
		config, err = rest.InClusterConfig()
	}
	checkError(err)

	// Create a client for the hiveV1 CustomResourceDefinitions
	hiveset, err := hiveclient.NewForConfig(config)
	checkError(err)

	// Create a client for kubernetes to access events
	kubeset, err := kubernetes.NewForConfig(config)
	checkError(err)

	podNamespace := os.Getenv("POD_NAMESPACE")

	// When running inside the cluster namespace as cluster-installer, we only have access to Get & Update for ClusterDeployment
	if podNamespace != "" {
		clusterDeployment, err := hiveset.HiveV1().ClusterDeployments(podNamespace).Get(context.TODO(), podNamespace, v1.GetOptions{})
		checkError(err)

		takeAction(hiveset, kubeset, *clusterDeployment, TakeAction, ClusterInstallerSA)
		fmt.Println("  \\-> Event supressed")

	} else {
		// Grab all ClusterDeployments to change the state
		cds, err := hiveset.HiveV1().ClusterDeployments(podNamespace).List(context.TODO(), v1.ListOptions{})
		checkError(err)

		for _, clusterDeployment := range cds.Items {

			if (OptIn == "true" && clusterDeployment.Labels["hibernate"] == "true") || (OptIn != "true" && clusterDeployment.Labels["hibernate"] != "skip") {
				takeAction(hiveset, kubeset, clusterDeployment, TakeAction, HibernateSA)
			} else {
				fmt.Println("Skip    : " + clusterDeployment.Name + "  (currently " + string(clusterDeployment.Spec.PowerState) + ")")
				fireEvent(kubeset, &clusterDeployment, "hibernating", "Skipping cluster "+clusterDeployment.Name, "skipAction", "Normal")
			}
		}
	}
}

func takeAction(hiveset *hiveclient.Clientset, kubeset *kubernetes.Clientset, clusterDeployment hivev1.ClusterDeployment, takeAction string, hibernateSA bool) {
	if string(clusterDeployment.Spec.PowerState) != takeAction {

		fmt.Print(takeAction + ": " + clusterDeployment.Name)

		var newPowerState string
		if hibernateSA {
			newPowerState = powerStatePatch(hiveset, &clusterDeployment, takeAction)
		} else {
			newPowerState = powerStateUpdate(hiveset, &clusterDeployment, takeAction)
		}

		// Check the new state and report a response
		if newPowerState == takeAction {
			fmt.Println("  ✓")

			if hibernateSA {
				fireEvent(kubeset, &clusterDeployment, "hibernating", "The cluster "+clusterDeployment.Name+" has powerState "+takeAction, takeAction, "Normal")
			}
		} else {
			fmt.Println("  X")

			if hibernateSA {
				fireEvent(kubeset, &clusterDeployment, "hibernating", "The cluster "+clusterDeployment.Name+" did not set powerState to "+takeAction, "failedHibernating", "Warning")
			}
		}
	} else {
		fmt.Println("Skip    : " + clusterDeployment.Name + "  (currently " + string(clusterDeployment.Spec.PowerState) + ")")

		if hibernateSA {
			fireEvent(kubeset, &clusterDeployment, "hibernating", "Skipping cluster "+clusterDeployment.Name+", requested powerState "+takeAction+" equals current powerState "+string(clusterDeployment.Spec.PowerState), "skipHibernating", "Normal")
		}
	}
}
