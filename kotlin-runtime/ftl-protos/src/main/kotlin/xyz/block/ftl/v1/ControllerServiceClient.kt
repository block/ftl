// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.ControllerService in xyz/block/ftl/v1/ftl.proto
package xyz.block.ftl.v1

import com.squareup.wire.GrpcCall
import com.squareup.wire.GrpcStreamingCall
import com.squareup.wire.Service

public interface ControllerServiceClient : Service {
  /**
   * Ping service for readiness.
   */
  public fun Ping(): GrpcCall<PingRequest, PingResponse>

  public fun Status(): GrpcCall<StatusRequest, StatusResponse>

  /**
   * Get list of artefacts that differ between the server and client.
   */
  public fun GetArtefactDiffs(): GrpcCall<GetArtefactDiffsRequest, GetArtefactDiffsResponse>

  /**
   * Upload an artefact to the server.
   */
  public fun UploadArtefact(): GrpcCall<UploadArtefactRequest, UploadArtefactResponse>

  /**
   * Create a deployment.
   */
  public fun CreateDeployment(): GrpcCall<CreateDeploymentRequest, CreateDeploymentResponse>

  /**
   * Get the schema and artefact metadata for a deployment.
   */
  public fun GetDeployment(): GrpcCall<GetDeploymentRequest, GetDeploymentResponse>

  /**
   * Stream deployment artefacts from the server.
   *
   * Each artefact is streamed one after the other as a sequence of max 1MB
   * chunks.
   */
  public fun GetDeploymentArtefacts():
      GrpcStreamingCall<GetDeploymentArtefactsRequest, GetDeploymentArtefactsResponse>

  /**
   * Register a Runner with the Controller.
   *
   * Each runner issue a RegisterRunnerRequest to the ControllerService
   * every 10 seconds to maintain its heartbeat.
   */
  public fun RegisterRunner(): GrpcStreamingCall<RunnerHeartbeat, RegisterRunnerResponse>

  /**
   * Update an existing deployment.
   */
  public fun UpdateDeploy(): GrpcCall<UpdateDeployRequest, UpdateDeployResponse>

  /**
   * Gradually replace an existing deployment with a new one.
   *
   * If a deployment already exists for the module of the new deployment,
   * it will be scaled down and replaced by the new one.
   */
  public fun ReplaceDeploy(): GrpcCall<ReplaceDeployRequest, ReplaceDeployResponse>

  /**
   * Stream logs from a deployment
   */
  public fun StreamDeploymentLogs():
      GrpcStreamingCall<StreamDeploymentLogsRequest, StreamDeploymentLogsResponse>

  /**
   * Pull schema changes from the Controller.
   */
  public fun PullSchema(): GrpcStreamingCall<PullSchemaRequest, PullSchemaResponse>
}
