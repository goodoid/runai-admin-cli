# Run:AI [![CircleCI](https://circleci.com/gh/run-ai/runai-cli.svg?style=svg)](https://circleci.com/gh/run-ai/runai-cli)
## Overview

Run:AI CLI is a command-line interface for the data scientists to run and monitor machine learning jobs on top of the Run:AI software and Kubernestes.

## Prerequisites
* Kubernetes 1.15+
* Kubectl installed and configured to access your cluster. Please refer to https://kubernetes.io/docs/tasks/tools/install-kubectl/
* Install Helm 3
* Run:AI software installed on your Kubernetes cluster. Please refer to https://support.run.ai/hc/en-us/articles/360010280179-Installing-Run-AI-on-an-on-premise-Kubernetes-Cluster for installation, if you haven't done so already.
## Setup

* Download latest release from the releases page. https://github.com/run-ai/arena/releases. 
* Unarchive the downloaded file.
* Install by running:
```
sudo ./install-runai.sh
```
To verify installation:
```
runai --help
```
## Quickstart

For help on Run:AI CLI run
```
runai --help
```
To verify the status of your cluster, use the `top` command.
```
runai top job
runai top node
```
These commands will give you valuable information about your cluster's and jobs' GPUs allocation status.

To run a sample job using runai sample training container please run:
```
runai submit -g 1 --name runai-test --project {your_project} -i gcr.io/run-ai-lab/quickstart -g 1
```
This will run a job using the Run:AI scheduler on top of the kubernetes cluster using Run:AI quickstart image and requesting 1 GPU for the job. To see the status of the running job please run:
```
runai list
```
Once the job in running, you can view its logs by running:
```
runai logs runai-test
```
At last, to delete the job prior to its completion you can run:
```
runai delete runai-test
```

## Updating the Run:AI CLI
To update the CLI to the latest version run:
```
sudo runai update
```
