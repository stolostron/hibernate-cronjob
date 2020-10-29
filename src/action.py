#author: jpacker@redhat.com
import kubernetes.client
import os
from kubernetes.client.rest import ApiException

if 'TAKE_ACTION' not in os.environ:
    raise EnvironmentError("Environment variable TAKE_ACTION missing")

TAKE_ACTION = os.environ['TAKE_ACTION']
if TAKE_ACTION != 'Hibernating' and TAKE_ACTION != 'Running':
    raise SyntaxError("TAKE_ACTION must be set to \"Hibernating\" or \"Running\", found: " + TAKE_ACTION)

configuration = kubernetes.client.Configuration()
configuration.verify_ssl = False

# Read API key Bearer Token 
CM_TOKEN = open('/var/run/secrets/kubernetes.io/serviceaccount/token', 'r').read()

configuration.api_key = {"authorization": "Bearer " + CM_TOKEN}
configuration.host = "https://kubernetes.default.svc.cluster.local"

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
            
            # Look for the ACM managed object
            managedCluster = api_instance.get_cluster_custom_object("cluster.open-cluster-management.io", "v1", "managedclusters",clusterName)
        
            if ('hibernate' in clusterObject['metadata']['labels'] and 'skip' == clusterObject['metadata']['labels']['hibernate']) or \
                    ('hibernate' in managedCluster['metadata']['labels'] and 'skip' == managedCluster['metadata']['labels']['hibernate']):

                print("Skip     : " + clusterName)

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
                api_instance.patch_namespaced_custom_object("hive.openshift.io", "v1", namespaceName,"clusterdeployments",clusterName, clusterPatch)
                # Now validate the change
                managedCluster = api_instance.get_namespaced_custom_object("hive.openshift.io", "v1", namespaceName, "clusterdeployments",clusterName)
                if managedCluster['spec']['powerState'] != TAKE_ACTION:
                    print('  X ')
                else:
                    print('  ✓')

