#author: jpacker@redhat.com
import kubernetes.client
import os
from datetime import datetime
from pprint import pprint
from kubernetes.client.rest import ApiException

# fire_event will create or update an existing event
def fire_event(clusterName, namespaceName, eventName, message, reason, eType, api_core):
    objRef = kubernetes.client.V1ObjectReference(kind='ClusterDeployment', name=clusterName, namespace=namespaceName)
    metaRef = kubernetes.client.V1ObjectMeta(name=eventName, namespace=namespaceName)

    body = kubernetes.client.V1Event(involved_object=objRef, metadata=metaRef)
    body.message = message
    body.reason = reason
    body.type = eType
    body.last_timestamp = datetime.utcnow().isoformat() + "Z"

    existingEvent = None
    try:  # Replace with the list_namespaced_events in the future
        existingEvent = api_core.read_namespaced_event(eventName, namespaceName)
    except:
        print("  \-> Event not found")

    if existingEvent:
        if not existingEvent.count:
            body.count = 1
        else:
            body.count = existingEvent.count + 1
        api_core.patch_namespaced_event(eventName, namespaceName, body)
        print("  \-> Update existing event " + eventName + " count " + str(body.count))
    else:
        body.count = 0
        api_core.create_namespaced_event(namespaceName, body)
        print("  \-> Create a new event " + eventName)

# Main code path
if 'TAKE_ACTION' not in os.environ:
    raise EnvironmentError("Environment variable TAKE_ACTION missing")

TAKE_ACTION = os.environ['TAKE_ACTION']
if TAKE_ACTION != 'Hibernating' and TAKE_ACTION != 'Running':
    raise SyntaxError("TAKE_ACTION must be set to \"Hibernating\" or \"Running\", found: " + TAKE_ACTION)

configuration = kubernetes.client.Configuration()
configuration.verify_ssl = False

# Read API key Bearer Token
CM_TOKEN = os.environ['CM_TOKEN']
if CM_TOKEN == "":
    print("Read token for Service Account from a secret")
    CM_TOKEN = open('/var/run/secrets/kubernetes.io/serviceaccount/token', 'r').read()
configuration.api_key = {"authorization": "Bearer " + CM_TOKEN}

# Read the API URL
CM_API_URL = os.environ['CM_API_URL']
if CM_API_URL == "":
    CM_API_URL = "https://kubernetes.default.svc.cluster.local"
configuration.host = CM_API_URL

with kubernetes.client.ApiClient(configuration) as api_client:
    # Create instances of the API class
    api_instance = kubernetes.client.CustomObjectsApi(api_client)
    api_core = kubernetes.client.CoreV1Api(api_client)

    # Get all the namespaces as there is no list_custom_object_for_all_namespaces
    for namespace in api_core.list_namespace().items:
        namespaceName = namespace.metadata.name

        # Query for the clusterDeployment kind
        api_response = api_instance.list_namespaced_custom_object("hive.openshift.io", "v1", namespaceName, "clusterdeployments")
        
        # Only process the namespace if a clusterDeployment is found
        if api_response['items'] != []:
            clusterObject = api_response['items'][0]
            clusterName = clusterObject['metadata']['name']
            
            # Look for the ACM managed object, may contain the hibernate=skip label
            managedCluster = api_instance.get_cluster_custom_object("cluster.open-cluster-management.io", "v1", "managedclusters",clusterName)

            if ('hibernate' in clusterObject['metadata']['labels'] and 'skip' == clusterObject['metadata']['labels']['hibernate']) or \
                    ('hibernate' in managedCluster['metadata']['labels'] and 'skip' == managedCluster['metadata']['labels']['hibernate']):

                print("Skip     : " + clusterName)
                fire_event(clusterName, namespaceName, "skiphibernating", "Skipping cluster " + clusterName + " labels.hibernate=skip. It will not be hibernating", "skipHibernating", "Normal", api_core)
            else:
                if clusterName != namespaceName:
                    print ("Skip     : Namespace: " + namespaceName + " does not match cluster name: " + clusterName + "")
                    continue

                print(TAKE_ACTION + ": " + clusterName, end='')
                clusterPatch = {
                    "apiVersion": "hive.openshift.io/v1",
                    "kind": "ClusterDeployment",
                    "metadata": {
                        "name": clusterName
                    },
                    "namespace": clusterName,
                    "spec": {
                        "powerState": TAKE_ACTION
                    }
                }

                # Hibernate the cluster
                api_instance.patch_namespaced_custom_object("hive.openshift.io", "v1", namespaceName,"clusterdeployments",clusterName, clusterPatch)

                # Now validate the powerState change
                managedCluster = api_instance.get_namespaced_custom_object("hive.openshift.io", "v1", namespaceName, "clusterdeployments",clusterName)
                if not 'powerState' in managedCluster['spec'] or managedCluster['spec']['powerState'] != TAKE_ACTION:
                    print('  X ')
                    fire_event(clusterName, namespaceName, "failedhibernating", "The cluster " + clusterName + " did not set powerState to Hibernating", "failedHibernating", "Warning", api_core)
                else:
                    print('  ✓')
                    fire_event(clusterName, namespaceName, "hibernating", "The cluster " + clusterName + " has powerState Hibernating", "Hibernating", "Normal", api_core)





