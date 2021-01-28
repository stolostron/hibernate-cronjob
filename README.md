# Hibernate your Hive provisioned clusters
This repository has the source code as well as the deployment for two Kubernetes CronJobs that will Hibernate and Resume your OpenShift clusters.

## Requirements
- Red Hat Advanced Cluster Management for Kubernetes 2.1
- OpenShift 4.5 and 4.6 clusters provisioned by ACM

## Deploy
1. Log into your ACM Hub on OpenShift and make sure `oc project -q` displays the name of the namespace where you want to install.
2. Run the command:
   ```
   make cronjobs
   ```
3. A ServiceAccount called `hibernator` was created. To grant the required privileges cluster-wide, run the command:
   ```
   make roles
   ```
   If you prefer to grant finer-grained privileges for specific namespaces, see [`templates/roles.yaml`](./templates/roles.yaml) for details.
4. Monitor the CronJobs
   ```
   oc get cronjobs
   ```

## Customizations

For a full list of options, run:
```
make params
```

To change any options, create or edit the `options.env` file. (NOTE: This file is not under version control.) For example, to disable automatic resumption of clusters, add the following to the `options.env` file:
```
RUNNING_DISABLED=true
```

### Time window
The default setting is to have the clusters running from 8am - 6pm Mon - Fri EST.  You can adjust the times by adding the `HIBERNATING_SCHEDULE` AND `RUNNING_SCHEDULE` params to `options.env`.

It uses the standard CronJob format and the check is done against a UTC clock.  For EDT, that means +4hrs.

```
#       Time is measured in coordinated universal time (UTC)
#                ┌───────────── minute (0 - 59)
#                │ ┌───────────── hour (0 - 23)
#                │ │ ┌───────────── day of the month (1 - 31)
#                │ │ │ ┌───────────── month (1 - 12)
#                │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday - Saturday)
RUNNING_SCHEDULE=0 13 * * 1-5"
```

### Skipping Clusters
The job will only target clusters deployed by Hive. It looks for the ClusterDeployment objects.  If you put a label on these objects `hibernate: skip` they will be ignored by both CronJobs.

For an opt-in model, instead specify `OPT_IN=true` in the `options.env` file. In this case, only ClusterDeployments with the label `hibernate: 'true'` will be acted upon.
## Runonce job

You can hibernate or resume all applicable clusters manually. These commands use the `options.env` file, so the value of the `OPT_IN` parameter is respected.
### To bring clusters back to "Ready"
```
make running
```
### To hibernate clusters
```
make hibernate
```

## Manual updates
Edit the ClusterDeployment resource for the cluster you want to change the state for.  Find the `spec.powerState` key and change the value to either: `Hibernating` or `Running`

## Building yourself
You'll need docker and a connection to a registry that your OpenShift can reach.  Export the environment variables `REPO_URL` with your own registry URL and `VERSION` the tag. See `make all` for the development related targets. After pushing an image, you can use your own image by setting the `CURATOR_IMAGE` parameter in `options.env`.

## Running local
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
