// @generated by protoc-gen-connect-es v1.4.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/ftl.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { AcquireLeaseRequest, AcquireLeaseResponse, CallRequest, CallResponse, CreateDeploymentRequest, CreateDeploymentResponse, DeployRequest, DeployResponse, GetArtefactDiffsRequest, GetArtefactDiffsResponse, GetConfigRequest, GetConfigResponse, GetDeploymentArtefactsRequest, GetDeploymentArtefactsResponse, GetDeploymentRequest, GetDeploymentResponse, GetSchemaRequest, GetSchemaResponse, GetSecretRequest, GetSecretResponse, ListConfigRequest, ListConfigResponse, ListSecretsRequest, ListSecretsResponse, ModuleContextRequest, ModuleContextResponse, PingRequest, PingResponse, ProcessListRequest, ProcessListResponse, PublishEventRequest, PublishEventResponse, PullSchemaRequest, PullSchemaResponse, RegisterRunnerRequest, RegisterRunnerResponse, ReplaceDeployRequest, ReplaceDeployResponse, ReserveRequest, ReserveResponse, ResetSubscriptionRequest, ResetSubscriptionResponse, SendFSMEventRequest, SendFSMEventResponse, SetConfigRequest, SetConfigResponse, SetSecretRequest, SetSecretResponse, StatusRequest, StatusResponse, StreamDeploymentLogsRequest, StreamDeploymentLogsResponse, TerminateRequest, UnsetConfigRequest, UnsetConfigResponse, UnsetSecretRequest, UnsetSecretResponse, UpdateDeployRequest, UpdateDeployResponse, UploadArtefactRequest, UploadArtefactResponse } from "./ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";

/**
 * VerbService is a common interface shared by multiple services for calling Verbs.
 *
 * @generated from service xyz.block.ftl.v1.VerbService
 */
export const VerbService = {
  typeName: "xyz.block.ftl.v1.VerbService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.VerbService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Issue a synchronous call to a Verb.
     *
     * @generated from rpc xyz.block.ftl.v1.VerbService.Call
     */
    call: {
      name: "Call",
      I: CallRequest,
      O: CallResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * ModuleService is the service that modules use to interact with the Controller.
 *
 * @generated from service xyz.block.ftl.v1.ModuleService
 */
export const ModuleService = {
  typeName: "xyz.block.ftl.v1.ModuleService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.ModuleService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Get configuration state for the module
     *
     * @generated from rpc xyz.block.ftl.v1.ModuleService.GetModuleContext
     */
    getModuleContext: {
      name: "GetModuleContext",
      I: ModuleContextRequest,
      O: ModuleContextResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Acquire (and renew) a lease for a deployment.
     *
     * Returns ResourceExhausted if the lease is held.
     *
     * @generated from rpc xyz.block.ftl.v1.ModuleService.AcquireLease
     */
    acquireLease: {
      name: "AcquireLease",
      I: AcquireLeaseRequest,
      O: AcquireLeaseResponse,
      kind: MethodKind.BiDiStreaming,
    },
    /**
     * Send an event to an FSM.
     *
     * @generated from rpc xyz.block.ftl.v1.ModuleService.SendFSMEvent
     */
    sendFSMEvent: {
      name: "SendFSMEvent",
      I: SendFSMEventRequest,
      O: SendFSMEventResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Set the next event for an FSM.
     *
     * @generated from rpc xyz.block.ftl.v1.ModuleService.SetNextFSMEvent
     */
    setNextFSMEvent: {
      name: "SetNextFSMEvent",
      I: SendFSMEventRequest,
      O: SendFSMEventResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Publish an event to a topic.
     *
     * @generated from rpc xyz.block.ftl.v1.ModuleService.PublishEvent
     */
    publishEvent: {
      name: "PublishEvent",
      I: PublishEventRequest,
      O: PublishEventResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * @generated from service xyz.block.ftl.v1.ControllerService
 */
export const ControllerService = {
  typeName: "xyz.block.ftl.v1.ControllerService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * List "processes" running on the cluster.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.ProcessList
     */
    processList: {
      name: "ProcessList",
      I: ProcessListRequest,
      O: ProcessListResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc xyz.block.ftl.v1.ControllerService.Status
     */
    status: {
      name: "Status",
      I: StatusRequest,
      O: StatusResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get list of artefacts that differ between the server and client.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.GetArtefactDiffs
     */
    getArtefactDiffs: {
      name: "GetArtefactDiffs",
      I: GetArtefactDiffsRequest,
      O: GetArtefactDiffsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Upload an artefact to the server.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.UploadArtefact
     */
    uploadArtefact: {
      name: "UploadArtefact",
      I: UploadArtefactRequest,
      O: UploadArtefactResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Create a deployment.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.CreateDeployment
     */
    createDeployment: {
      name: "CreateDeployment",
      I: CreateDeploymentRequest,
      O: CreateDeploymentResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get the schema and artefact metadata for a deployment.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.GetDeployment
     */
    getDeployment: {
      name: "GetDeployment",
      I: GetDeploymentRequest,
      O: GetDeploymentResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Stream deployment artefacts from the server.
     *
     * Each artefact is streamed one after the other as a sequence of max 1MB
     * chunks.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.GetDeploymentArtefacts
     */
    getDeploymentArtefacts: {
      name: "GetDeploymentArtefacts",
      I: GetDeploymentArtefactsRequest,
      O: GetDeploymentArtefactsResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Register a Runner with the Controller.
     *
     * Each runner issue a RegisterRunnerRequest to the ControllerService
     * every 10 seconds to maintain its heartbeat.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.RegisterRunner
     */
    registerRunner: {
      name: "RegisterRunner",
      I: RegisterRunnerRequest,
      O: RegisterRunnerResponse,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * Update an existing deployment.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.UpdateDeploy
     */
    updateDeploy: {
      name: "UpdateDeploy",
      I: UpdateDeployRequest,
      O: UpdateDeployResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Gradually replace an existing deployment with a new one.
     *
     * If a deployment already exists for the module of the new deployment,
     * it will be scaled down and replaced by the new one.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.ReplaceDeploy
     */
    replaceDeploy: {
      name: "ReplaceDeploy",
      I: ReplaceDeployRequest,
      O: ReplaceDeployResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Stream logs from a deployment
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.StreamDeploymentLogs
     */
    streamDeploymentLogs: {
      name: "StreamDeploymentLogs",
      I: StreamDeploymentLogsRequest,
      O: StreamDeploymentLogsResponse,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * Get the full schema.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.GetSchema
     */
    getSchema: {
      name: "GetSchema",
      I: GetSchemaRequest,
      O: GetSchemaResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Pull schema changes from the Controller.
     *
     * Note that if there are no deployments this will block indefinitely, making it unsuitable for
     * just retrieving the schema. Use GetSchema for that.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.PullSchema
     */
    pullSchema: {
      name: "PullSchema",
      I: PullSchemaRequest,
      O: PullSchemaResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Reset the cursor for a subscription to the head of its topic.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.ResetSubscription
     */
    resetSubscription: {
      name: "ResetSubscription",
      I: ResetSubscriptionRequest,
      O: ResetSubscriptionResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * RunnerService is the service that executes Deployments.
 *
 * The Controller will scale the Runner horizontally as required. The Runner will
 * register itself automatically with the ControllerService, which will then
 * assign modules to it.
 *
 * @generated from service xyz.block.ftl.v1.RunnerService
 */
export const RunnerService = {
  typeName: "xyz.block.ftl.v1.RunnerService",
  methods: {
    /**
     * @generated from rpc xyz.block.ftl.v1.RunnerService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Reserve synchronously reserves a Runner for a deployment but does nothing else.
     *
     * @generated from rpc xyz.block.ftl.v1.RunnerService.Reserve
     */
    reserve: {
      name: "Reserve",
      I: ReserveRequest,
      O: ReserveResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Initiate a deployment on this Runner.
     *
     * @generated from rpc xyz.block.ftl.v1.RunnerService.Deploy
     */
    deploy: {
      name: "Deploy",
      I: DeployRequest,
      O: DeployResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Terminate the deployment on this Runner.
     *
     * @generated from rpc xyz.block.ftl.v1.RunnerService.Terminate
     */
    terminate: {
      name: "Terminate",
      I: TerminateRequest,
      O: RegisterRunnerRequest,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * AdminService is the service that provides and updates admin data. For example,
 * it is used to encapsulate configuration and secrets.
 *
 * @generated from service xyz.block.ftl.v1.AdminService
 */
export const AdminService = {
  typeName: "xyz.block.ftl.v1.AdminService",
  methods: {
    /**
     * @generated from rpc xyz.block.ftl.v1.AdminService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * List configuration.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.ConfigList
     */
    configList: {
      name: "ConfigList",
      I: ListConfigRequest,
      O: ListConfigResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get a config value.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.ConfigGet
     */
    configGet: {
      name: "ConfigGet",
      I: GetConfigRequest,
      O: GetConfigResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Set a config value.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.ConfigSet
     */
    configSet: {
      name: "ConfigSet",
      I: SetConfigRequest,
      O: SetConfigResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Unset a config value.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.ConfigUnset
     */
    configUnset: {
      name: "ConfigUnset",
      I: UnsetConfigRequest,
      O: UnsetConfigResponse,
      kind: MethodKind.Unary,
    },
    /**
     * List secrets.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.SecretsList
     */
    secretsList: {
      name: "SecretsList",
      I: ListSecretsRequest,
      O: ListSecretsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get a secret.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.SecretGet
     */
    secretGet: {
      name: "SecretGet",
      I: GetSecretRequest,
      O: GetSecretResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Set a secret.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.SecretSet
     */
    secretSet: {
      name: "SecretSet",
      I: SetSecretRequest,
      O: SetSecretResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Unset a secret.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.SecretUnset
     */
    secretUnset: {
      name: "SecretUnset",
      I: UnsetSecretRequest,
      O: UnsetSecretResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

