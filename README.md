# Hibernate your Hive provisioned clusters and scale cluster pools
This repository has the source code as well as the deployment for two Kubernetes CronJobs that will Hibernate and Resume your OpenShift clusters.

In addition, there is tooling to create CronJobs for scaling cluster pools.

## How this repository can be used to hibernate Hive provisioned clusters
There are three supported configurations you can use:
1) Opt IN: A cluster will only be affected by the cronjob if it has the label `hibernate: true`
2) Opt OUT: A cluster will ALWAYS be affected by the cronjob, unless it has the label `hibernate: false`
3) Cluster Scoped: The cluster in the namespace where the cronjob is created is the only one affected by the ACTION value you choose (`hibernating` or `running`)

## Requirements
- Red Hat Advanced Cluster Management for Kubernetes v2.1+
- OpenShift `oc` command line tool installed
- OpenShift 4.5+ clusters provisioned by ACM
- Add the `NAMEPSPACE` parameter, pointing to where you want to run the hibernation jobs, in the `./options.env` file (Opt IN/OUT)
```bash
# ./options.env
NAMESPACE: open-cluster-management  #This can be any namespace where you want to install for Opt IN or Opt OUT
```

## Deploy Opt IN or Opt OUT scenarios
1. Log into your ACM Hub on OpenShift
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

## Deploy cluster scoped scenario
1. Log into your ACM Hub on OpenShift
2. Switch to the namespace(project) of the cluster you want to apply the cronjob to
2. Run the command to create cronjob
   ```bash
   # ACTION=hibernating   or  ACTION=running
   # SCHEDULE=<crontab_format>
   # NAMESPACE=<my-cluster-name>  The cluster name you wish to apply the cronjob to

   # Hibernating cronjob
   oc process -f templates/cluster-scoped-cronjob.yaml ACTION=hibernating SCHEDULE="0 23 * * 0-6" NAMESPACE= my-cluster-name --ignore-unknown-parameters=true  | oc apply -f -

   # Running cronjob
   oc process -f templates/cluster-scoped-cronjob.yaml ACTION=running SCHEDULE="0 12 * * 0-6" NAMESPACE= my-cluster-name --ignore-unknown-parameters=true  | oc apply -f -
   ```
   _Note: The cluster scoped cronjob requires the value `hibernating` or `running` to be lowercase. This is becuase the `ACTION` parameter is also used to generate the cronjob name.

## Customizations for Opt IN & OUT

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
## Run once job

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

# Cluster pool scaling
Use Kubernetes CronJobs to scale up or down cluster pools at specific times.  Reducing the size of a cluster pool during off-peak hours can help reduce cloud expense.

## Customizations

For a full list of options, run:
```
make clusterpool-params
```

NOTE: This feature uses the ServiceAccount called `hibernator` that was created previously in the section above.

For example, to scale up a cluster pool named `aws-4-10`, create a file `aws-4-10-scaleup.env` containing the following options. (NOTE: This file is not under version control.) For example, to change the pool size at 8am Mon - Fri EST:
```
CRONJOB_SCHEDULE: 0 12 * * 1-5
SERVICE_ACCOUNT_NAME: hibernator
CLUSTER_POOL_NAME: aws-4-10
CLUSTER_POOL_SIZE: 4
NAME: aws-4-10-scaleup
NAMESPACE: my-namespace
```

For a complementary scale down of a cluster pool, create a file `aws-4-10-scaledown.env` containing the following options. (NOTE: This file is not under version control.) For example, to change the pool size at midnight Tue - Sat EST:
```
CRONJOB_SCHEDULE: 0 4 * * 2-6
SERVICE_ACCOUNT_NAME: hibernator
CLUSTER_POOL_NAME: aws-4-10
CLUSTER_POOL_SIZE: 2
NAME: aws-4-10-scaledown
NAMESPACE: my-namespace
```


## Deploy the CronJob
Run the following command to create the CronJob to sclae up the cluster pool:
```
oc process -f templates/scale-clusterpool-job.yaml --param-file ./aws-4-10-scaleup.env  --ignore-unknown-parameters=true | oc apply -f -
```

Run the following command to create the CronJob to sclae down the cluster pool:
```
oc process -f templates/scale-clusterpool-job.yaml --param-file ./aws-4-10-scaledown.env  --ignore-unknown-parameters=true | oc apply -f -
```

## View CronJobs
1. Log into your ACM Hub on OpenShift
2. Monitor the CronJobs
   ```
   oc get cronjobs -n <NAMESPACE>
   ```

## Manual updates
Edit the ClusterPool resource for the cluster you want to change the size for.  Find the `spec.size` key and change the value to the size you want.
