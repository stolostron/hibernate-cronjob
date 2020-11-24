#author: github.com/jnpacker
import kubernetes.client
from datetime import datetime

# fire_event will create or update an existing event
def fire(objName, namespaceName, objKind, eventName, message, reason, eType, api_core):
    objRef = kubernetes.client.V1ObjectReference(kind=objKind, name=objName, namespace=namespaceName)
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
        print("  |-> Event not found")

    if existingEvent:
        if not existingEvent.count:
            body.count = 1
        else:
            body.count = existingEvent.count + 1
        api_core.patch_namespaced_event(eventName, namespaceName, body)
        print("  \\-> Update existing event " + eventName + ", event count " + str(body.count))
    else:
        body.count = 1
        api_core.create_namespaced_event(namespaceName, body)
        print("  \\-> Create a new event " + eventName)