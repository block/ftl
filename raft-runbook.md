# Raft Cluster Management Runbook

This document describes operational procedures for managing the Raft cluster, including scaling operations and disaster recovery.

## Architecture Overview

The Raft cluster consists of:
- Multiple stateful pods running in Kubernetes
- Each pod maintains its own copy of the Raft state in its persistent volume
- Cluster membership is managed through a control API
- Each node has a unique replica ID stored in `replica_id` file in its data directory. A replica_id can not be reused once it has been decomissioned

## Scaling Operations

### Scaling Up

1. Increase the number of replicas in the StatefulSet:
   ```bash
   kubectl scale statefulset ftl-schema --replicas=<new-count>
   ```

2. The new pods will automatically:
   - Generate a unique replica ID
   - Contact the control plane to join the cluster
   - Begin replicating data from existing members

3. Monitor the cluster health:
   ```bash
   kubectl logs -f ftl-schema-<new-count>
   ```
   Look for messages indicating successful join and replication.

### Scaling Down

1. Before scaling down, ensure the cluster has enough replicas to maintain quorum:
   - At least 3 nodes

2. Remove the member from the Raft cluster first:
   ```bash
   curl --header "Content-Type: application/json" --request POST --data '{"replica_id": <replica id>, "shard_ids": [<shard ids>]}' <host>/xyz.block.ftl.raft.v1.RaftService/RemoveMember
   ```
   You need to be able to call the control API for this. You can call any of the raft group pods.
   You will need to include all the shard ids that the node is a part of in the request.

3. Scale down the StatefulSet:
   ```bash
   kubectl scale statefulset ftl-schema --replicas=<new-count>
   ```

4. Remove the PVC to ensure clean state:
   ```bash
   kubectl delete pvc <pvc-name>
   ```

## Disaster Recovery

### Recovering from Corrupted Data

If a node's data becomes corrupted:

1. Remove the corrupted member from the cluster:
   ```bash
   curl --header "Content-Type: application/json" --request POST --data '{"replica_id": <replica id>, "shard_ids": [<shard ids>]}' <host>/xyz.block.ftl.raft.v1.RaftService/RemoveMember
   ```

2. If the node is an initial member, you will need to change the replica id for that member to a new value in the startup parameters.

3. Delete the pod:
   ```bash
   kubectl delete pod ftl-schema-<replica id>
   ```

4. Delete the PVC to ensure clean state:
   ```bash
   kubectl delete pvc <pvc-name>
   ```

5. The StatefulSet controller will:
   - Create a new PVC
   - Restart the pod
   - The pod will join as a new member with fresh state

### Recovering from Quorum Loss

TODO
