// @generated by protoc-gen-connect-es v0.9.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/ftl.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { CallRequest, CallResponse, CreateDeploymentRequest, CreateDeploymentResponse, DeployRequest, DeployResponse, DeployToRunnerRequest, DeployToRunnerResponse, GetArtefactDiffsRequest, GetArtefactDiffsResponse, GetDeploymentArtefactsRequest, GetDeploymentArtefactsResponse, GetDeploymentRequest, GetDeploymentResponse, ListRequest, ListResponse, PingRequest, PingResponse, PullSchemaRequest, PullSchemaResponse, PushSchemaRequest, PushSchemaResponse, RegisterRunnerRequest, RegisterRunnerResponse, SendMetricsRequest, SendMetricsResponse, StreamDeploymentLogsRequest, StreamDeploymentLogsResponse, UploadArtefactRequest, UploadArtefactResponse } from "./ftl_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

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
    /**
     * List the available Verbs.
     *
     * @generated from rpc xyz.block.ftl.v1.VerbService.List
     */
    list: {
      name: "List",
      I: ListRequest,
      O: ListResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * DevelService is the service that provides language-specific development and
 * deployment functionality.
 *
 * The DevelService is responsible for hot reloading when code changes, and
 * passing Verb calls through.
 *
 * @generated from service xyz.block.ftl.v1.DevelService
 */
export const DevelService = {
  typeName: "xyz.block.ftl.v1.DevelService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.DevelService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Push schema changes to the server.
     *
     * @generated from rpc xyz.block.ftl.v1.DevelService.PushSchema
     */
    pushSchema: {
      name: "PushSchema",
      I: PushSchemaRequest,
      O: PushSchemaResponse,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * Pull schema changes from the server.
     *
     * @generated from rpc xyz.block.ftl.v1.DevelService.PullSchema
     */
    pullSchema: {
      name: "PullSchema",
      I: PullSchemaRequest,
      O: PullSchemaResponse,
      kind: MethodKind.ServerStreaming,
    },
  }
} as const;

/**
 * @generated from service xyz.block.ftl.v1.ControlPlaneService
 */
export const ControlPlaneService = {
  typeName: "xyz.block.ftl.v1.ControlPlaneService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get list of artefacts that differ between the server and client.
     *
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.GetArtefactDiffs
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
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.UploadArtefact
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
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.CreateDeployment
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
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.GetDeployment
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
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.GetDeploymentArtefacts
     */
    getDeploymentArtefacts: {
      name: "GetDeploymentArtefacts",
      I: GetDeploymentArtefactsRequest,
      O: GetDeploymentArtefactsResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Register a Runner with the ControlPlane.
     *
     * Each runner MUST stream a RegisterRunnerRequest to the ControlPlaneService
     * every 10 seconds to maintain its heartbeat.
     *
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.RegisterRunner
     */
    registerRunner: {
      name: "RegisterRunner",
      I: RegisterRunnerRequest,
      O: RegisterRunnerResponse,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * Starts a deployment.
     *
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.Deploy
     */
    deploy: {
      name: "Deploy",
      I: DeployRequest,
      O: DeployResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Stream logs from a deployment
     *
     * @generated from rpc xyz.block.ftl.v1.ControlPlaneService.StreamDeploymentLogs
     */
    streamDeploymentLogs: {
      name: "StreamDeploymentLogs",
      I: StreamDeploymentLogsRequest,
      O: StreamDeploymentLogsResponse,
      kind: MethodKind.ClientStreaming,
    },
  }
} as const;

/**
 * RunnerService is the service that executes Deployments.
 *
 * The ControlPlane will scale the Runner horizontally as required. The Runner will
 * register itself automatically with the ControlPlaneService, which will then
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
    },
    /**
     * Initiate a deployment on this Runner.
     *
     * @generated from rpc xyz.block.ftl.v1.RunnerService.DeployToRunner
     */
    deployToRunner: {
      name: "DeployToRunner",
      I: DeployToRunnerRequest,
      O: DeployToRunnerResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * @generated from service xyz.block.ftl.v1.ObservabilityService
 */
export const ObservabilityService = {
  typeName: "xyz.block.ftl.v1.ObservabilityService",
  methods: {
    /**
     * @generated from rpc xyz.block.ftl.v1.ObservabilityService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Send OTEL metrics from the Deployment to the ControlPlane via the Runner.
     *
     * @generated from rpc xyz.block.ftl.v1.ObservabilityService.SendMetrics
     */
    sendMetrics: {
      name: "SendMetrics",
      I: SendMetricsRequest,
      O: SendMetricsResponse,
      kind: MethodKind.ClientStreaming,
    },
  }
} as const;

