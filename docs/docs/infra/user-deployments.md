---
sidebar_position: 1
title: User Deployments
description: User Deployments
---

# User Deployments

In FTL user deployments are started via an FTL runner process, that launches the user code and acts as a proxy to the outside world.

A user deployment is given a unique ID at deployment time, which is the deployment name. This will include the module name and realm, and a unique suffix.

In a kubernetes environment the FTL provisioner will create a new Kubernetes deployments that uses the FTL runner image. When this image starts
it will start the runner process, which connects to the `ftl-controller` service to get the deployment configuration.

Part of this configuration include the digests of the artifacts that make up the deployment, once it has these artifacts the runner
will download them from an OCI registry and lay them out in a local directory. The runner will then start the user code,
passing it the address of the various services the runner is proxying. 


