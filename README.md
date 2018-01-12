# OPENSHIFT JOB CLEANER

[![pipeline status](https://gitlab.nine.ch/ninech/openshift-job-cleaner/badges/master/pipeline.svg)](https://gitlab.nine.ch/ninech/openshift-job-cleaner/commits/master)
[![coverage report](https://gitlab.nine.ch/ninech/openshift-job-cleaner/badges/master/coverage.svg)](https://gitlab.nine.ch/ninech/openshift-job-cleaner/commits/master)

A basic openshift job cleaner for your cluster.
Idea from https://github.com/willemvd/openshift-scheduledjobs-cleanup but refactored as a golang application and
provides additional configuration and logging.

## Installation

```
oc new-project cleaner
oc create -f openshift/openshift-job-cleaner.list.yml
```
This will create the objects you need and a default configuration map.

You will then need to configure your serviceaccount to have access to the projects it should be able to clean. It is recommended to give
the sa a cluster wide edit role and use the `blacklist` configuration to exclude unwanted namespaces.
```
oc adm policy add-cluster-role-to-user edit system:serviceaccount:cleaner:openshift-job-cleaner
```

## Configuration

### Application Config

See [example config](openshift/openshift-job-clearer.config.example.yml) for an example configuration.
- max-age can be specified as default values or as namespace specific values
- max-age for success and failure can be configured differently
- a job is not considered failed unless it reaches it's `activeDeadlineSeconds`

max-age is in MINUTES.

#### Blacklist
Any namespace in the blacklist is excluded by the job cleaner for both successful and failed jobs

#### Max-age
- For successful jobs the max-age is taken from completion
- For failed jobs the max-age is taken from termination (and thus is affected by overall job timeout)

### Container config

#### ENV

The application is configured using environmental variables

- SENTRY_DSN                        - dns url for sentry, as standard for sentry based applications
- OJC_CONFIG_PATH                   - the path to the ocj.yml config file (including filename) if not set defaults to /opt/ocj/ocj.yml
- KUBERNETES_PORT_443_TCP_ADDR      - url to kube endpoint, set automatically
- KUBERNETES_SERVICE_PORT_HTTPS     - the port to use for the kube endpoint, set automatically
- KUBERNETES_SA_TOKEN_PATH          - override the default token path
- KUBERNETES_CERT_AUTHORITY_PATH    - override the default cert authority path

#### Secrets

The following secrets should be mounted in the container, usually this is handled automatically by openshift

- /var/run/secrets/kubernetes.io/serviceaccount/token - This is used as the token when logging in to oc as the service account user
- /var/run/secrets/kubernetes.io/serviceaccount/ca.crt - This is used as the certificate-authority for oc login

#### Logging

As you probably want to have good logging for a service that runs over all your namespaces this project uses logrus
which allows you to provide your own adapters and easily integrate logging systems without extensive code changes.

Currently it is only configured to use sentry but this can be easily altered from `main.go`

## Templates

the `openshift` folder contains openshift yml files that can be used to create and run jobs

### openshift-job-cleaner.list.yml
Creates the ServiceAccount, ConfigMap and CronJob needed for the service.
Does NOT give permissions to the ServiceAccount to access other namespaces, this should be done manually or added at a cluster level
```
oc new-project cleaner
oc create -f openshift/openshift-job-cleaner.list.yml
oc adm policy add-cluster-role-to-user edit system:serviceaccount:cleaner:openshift-job-cleaner
```

### openshift-job-cleaner.job.yml
One off job execution with dynamic name, should be used in the namespace with the ServiceAccount configured
```
oc project cleaner
oc create -f openshift/openshift-job-cleaner.job.yml
```

### openshift-job-clearer.config.example.yml
Example appliation configuration file showing all usable keys and with annotations

## Development/Testing

For development you will need a few tools and dependencies installed as well as docker and oc

```
go get -u github.com/golang/dep/cmd/dep
go get -u github.com/go-task/task/cmd/task
go get -u github.com/alecthomas/gometalinter
go get -u github.com/haya14busa/goverage

dep ensure

# NON GOLANG DEPENDENCIES

# See https://github.com/aelsabbahy/goss/releases for release versions
curl -L https://github.com/aelsabbahy/goss/releases/download/_VERSION_/goss-linux-amd64 -o /usr/local/bin/goss
chmod +rx /usr/local/bin/goss
#https://github.com/hadolint/hadolint/releases
brew install hadolint
```
This should allow you to run all of the tasks in the taskfile successfully.

The task `test:integration` will take ~15 minutes to complete and involves starting and stopping a local oc cluster