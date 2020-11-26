# Hibernate your Hive provisioned clusters
This repository has the source code as well as the deployment for two Kubernetes CronJobs that will Hibernate and Resume your OpenShift clusters.

## Requirements
- Red Hat Advanced Cluster Management for Kubernetes 2.1
- OpenShift 4.5 and 4.6 clusters provisioned by ACM

## Deploy
1. Log into your ACM Hub on OpenShift
2. Run the command:
```
make setup-go
```
4. Monitor the CornJobs
```
oc get cronjobs
```

## Time window
The default settings is to have the clusters runnings from 8am - 6pm Mon - Fri EST.  You can adjust the times by modifying the two CronJobs:
```
deploy/hibernating-cronjob.yaml
deploy/running-cronjob.yaml
```
It uses the standard CronJob format and the check is done against a UTC clock.  For EDT, that means +4hrs.

## Skipping Clusters
The job will only target clusters deployed by Hive. It looks for the ClusterDeployment objects.  If you put a label on either of these objects `hibernate: skip` they will be ignored by both CronJobs.

## Runonce job
### To bring clusters back to "Ready"
```
make running-go

## OR

oc create -f deploy-go/hibernation-job.yaml
```
### To hibernate clusters
```
make hibernate-go

## OR

oc create -f deploy-go/running-job.yaml
```

## Manual updates
Edit the ClusterDeployment resource for the cluster you want to change the state for.  Find the `spec.powerState` key and change the value to either: `Hibernating` or `Running`

## Building yourself
You'll need docker and a connection to a registry that your OpenShift can reach.  Modify the Makefile and replace the value of `REPO_URL` with your own registry URL. Next modify the two CronJobs to use your new image:
```
deploy-go/hibernating-cronjob.yaml
deploy-go/running-cronjob.yaml
```

## Running local (Python)
Setup the following environment variables, then run `python hibernate-cronjob/action.py`
```bash
export TARGET_ACTION=hibernating  # Or "running"
export CM_TOKEN=                  # Your OpenShift API token
export CM_API_URL                 # API URL for your cluster https://my.cluster.hostname.com:6443
```

## Running local (Python)
Setup the following environment variables, then run `go run ./pkg`
```bash
export REPO_URL=             # The repository you will upload your image too
export VERSION=              # The image version you want to use
```

## Time chart
| Time Zone | UTC offset | Example |
| :-------: | :--------: | :-----: |
| EST       | +5hr       | 11:22EST = 16:22UTC |
| EDT       | +4hr       | 11:22EDT = 15:22UTC |
| PST       | +8hr       | 11:22PST = 19:22UTC |

## Notes
* Added to the open-cluster-management community [2020-11-25](https://github.com/open-cluster-management/community/issues/6)
