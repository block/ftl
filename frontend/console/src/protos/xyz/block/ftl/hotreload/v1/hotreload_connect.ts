// @generated by protoc-gen-connect-es v1.6.1 with parameter "target=ts"
// @generated from file xyz/block/ftl/hotreload/v1/hotreload.proto (package xyz.block.ftl.hotreload.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "../../v1/ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { ReloadRequest, ReloadResponse, WatchRequest, WatchResponse } from "./hotreload_pb.js";

/**
 * HotReloadService is for communication between a language plugin a language runtime that can perform a hot reload
 *
 * @generated from service xyz.block.ftl.hotreload.v1.HotReloadService
 */
export const HotReloadService = {
  typeName: "xyz.block.ftl.hotreload.v1.HotReloadService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.hotreload.v1.HotReloadService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Forces an explicit Reload from the plugin. This is useful for when the plugin needs to trigger a Reload,
     * such as when the Reload context changes.
     *
     *
     * @generated from rpc xyz.block.ftl.hotreload.v1.HotReloadService.Reload
     */
    reload: {
      name: "Reload",
      I: ReloadRequest,
      O: ReloadResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Watch for hot reloads not initiated by an explicit Reload call.
     *
     * @generated from rpc xyz.block.ftl.hotreload.v1.HotReloadService.Watch
     */
    watch: {
      name: "Watch",
      I: WatchRequest,
      O: WatchResponse,
      kind: MethodKind.ServerStreaming,
    },
  }
} as const;

