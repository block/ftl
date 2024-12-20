// @generated by protoc-gen-connect-es v1.6.1 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/verb.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "./ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { CallRequest, CallResponse } from "./verb_pb.js";

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

