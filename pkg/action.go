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
	corev1 "k8s.io/api/core/v1"
	eventv1beta1 "k8s.io/api/events/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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
func checkError(err error) {
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
	checkError(err)
	return string(changedCD.Spec.PowerState)
}

// Used to create events for Cluster hibernation actions
//objName, namespaceName, objKind, eventName, message, reason, eType, api_core
func fireEvent(clientSet *kubernetes.Clientset, objRef *hivev1.ClusterDeployment, eventName string, message string, reason string, eType string) {
	event, err := clientSet.EventsV1beta1().Events(objRef.Namespace).Get(context.TODO(), eventName, v1.GetOptions{})
	if event != nil && event.Series == nil {
		event.Series = new(eventv1beta1.EventSeries)
		event.Series.Count = 1
		event.Series.LastObservedTime = v1.NowMicro()
	}
	if err != nil {
		fmt.Println("  |-> Event not found")
		event = new(eventv1beta1.Event)
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
		_, err := clientSet.EventsV1beta1().Events(objRef.Namespace).Create(context.TODO(), event, v1.CreateOptions{})
		fmt.Println("  \\-> Create a new event " + eventName + " for cluster " + objRef.Namespace + "/" + objRef.Name)
		checkError(err)
	} else {
		_, err := clientSet.EventsV1beta1().Events(objRef.Namespace).Update(context.TODO(), event, v1.UpdateOptions{})
		fmt.Println("  \\-> Update existing event "+eventName+", event count", event.Series.Count)
		checkError(err)
	}
}

func main() {
	var kubeconfig *string

	// Determine what action to take Hibernating or Running
	var TakeAction = os.Getenv("TAKE_ACTION")
	if TakeAction == "" || (TakeAction != "Hibernating" && TakeAction != "Running") {
		panic("Environment variable TAKE_ACTION missing: " + TakeAction)
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
	checkError(err)

	// Create a client for the hiveV1 CustomResourceDefinitions
	hiveset, err := hiveclient.NewForConfig(config)
	checkError(err)

	// Create a client for kubernetes to access events
	kubeset, err := kubernetes.NewForConfig(config)
	checkError(err)

	// Grab all ClusterDeployments to change the state
	cds, err := hiveset.HiveV1().ClusterDeployments("").List(context.TODO(), v1.ListOptions{})
	for _, clusterDeployment := range cds.Items {
		if clusterDeployment.Labels["hibernate"] != "skip" {
			if string(clusterDeployment.Spec.PowerState) != TakeAction {

				fmt.Print(TakeAction + ": " + clusterDeployment.Name)
				newPowerState := powerStateChange(hiveset, &clusterDeployment, TakeAction)

				// Check the new state and report a response
				if newPowerState == TakeAction {
					fmt.Println("  ✓")
					fireEvent(kubeset, &clusterDeployment, "hibernating", "The cluster "+clusterDeployment.Name+" has powerState "+TakeAction, TakeAction, "Normal")
				} else {
					fmt.Println("  X")
					//"failedhibernating", "The cluster " + clusterName + " did not set powerState to Hibernating", "failedHibernating", "Warning"
					fireEvent(kubeset, &clusterDeployment, "hibernating", "The cluster "+clusterDeployment.Name+" did not set powerState to "+TakeAction, "failedHibernating", "Warning")
				}
			} else {
				fmt.Println("Skip    : " + clusterDeployment.Name + "  (currently " + string(clusterDeployment.Spec.PowerState) + ")")
				fireEvent(kubeset, &clusterDeployment, "hibernating", "Skipping cluster "+clusterDeployment.Name+", requested powerState "+TakeAction+" equals current powerState "+string(clusterDeployment.Spec.PowerState), "skipHibernating", "Normal")
			}
		} else {
			fmt.Println("Skip    : " + clusterDeployment.Name + "  (currently " + string(clusterDeployment.Spec.PowerState) + ")")
			fireEvent(kubeset, &clusterDeployment, "hibernating", "Skipping cluster "+clusterDeployment.Name+", found labels.hibernate=skip. It will not be hibernating", "skipHibernating", "Normal")
		}
	}
}
