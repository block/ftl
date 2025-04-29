---
sidebar_position: 1
title: Kubernetes
description: Description of Kubernetes deployment infrastructure
---

# Kubernetes

FTL is designed to be deployed on Kubernetes, ideally with Istio installed for traffic management. This document describes the Kubernetes deployment infrastructure, including the use of Istio and other components.

## Components and Namespaces

By default, FTL will be deployed in the `ftl` namespace. The following deployments are included:

- Schema Service* (`ftl-schema`)
- Console (`ftl-console`)
- Admin Service (`ftl-admin`)
- Provisioner (`ftl-provisioner`)
- Cron Service* (`ftl-cron`)
- Lease Service (`ftl-lease`)
- Timeline Service (`ftl-timeline`)
- Ingress Service (`ftl-ingress`)

*StatefulSet with 3 replicas and Persistent Volumes

End user deployments will generally be deployed into either a separate shared namespace, or to a namespace per module
of the form <module>-<realm>. 


## Provisioning User Deployments

When a user deploys a module, the FTL Provisioner will create a new namespace if required for the module and deploy the module into that namespace. 
The namespace will be named `<module>-<realm>`, where `<module>` is the name of the module and `<realm>` is the name of the realm.

Each new version of the module gets a new kubernetes deployment, so that the old version can be kept running until the new version is verified to be working.
This will eventually allow FTL to perform canary deployments and other advanced deployment strategies. Each deployment has a service created for it, and
FTL controls the routing of traffic to the correct service. 

When a module is first deployed a Kube ServiceAccount will also be created with the same name as the module. This ServiceAccount will be used to run the module's pods, and is used as the workload identity.

If istio is installed in the cluster, the namespace will be provisioned with Istio sidecar injection enabled. When the modules are provisioned they will also have
Istio `AuthorizationPolicy` resources created to control the traffic to and from the module. FTL will provision these policies with the following rules:

- If the module has HTTP ingres verbs then it will allow traffic from the ingress service to the module
- If the module has Cron jobs then it will allow traffic from the cron service to the module
- Invocations from the `ftl-admin` service will be allowed for CLI invocations
- Invocations from `ftl-console` will be allowed for console invocations
- A new `AuthorizationPolicy` will be created for each module that is called by this module, as identified by the `+calls` metadata in the schema.

Other invocations are disallowed by default. This is to ensure that modules are not able to call each other without this being explicitly identified in the schema.


