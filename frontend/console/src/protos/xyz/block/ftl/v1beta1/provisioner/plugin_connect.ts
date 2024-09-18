// @generated by protoc-gen-connect-es v1.4.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1beta1/provisioner/plugin.proto (package xyz.block.ftl.v1beta1.provisioner, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "../../v1/ftl_pb.js";
import { MethodKind } from "@bufbuild/protobuf";
import { PlanRequest, PlanResponse, ProvisionRequest, ProvisionResponse, StatusRequest, StatusResponse } from "./plugin_pb.js";

/**
 * @generated from service xyz.block.ftl.v1beta1.provisioner.ProvisionerPluginService
 */
export const ProvisionerPluginService = {
  typeName: "xyz.block.ftl.v1beta1.provisioner.ProvisionerPluginService",
  methods: {
    /**
     * @generated from rpc xyz.block.ftl.v1beta1.provisioner.ProvisionerPluginService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc xyz.block.ftl.v1beta1.provisioner.ProvisionerPluginService.Provision
     */
    provision: {
      name: "Provision",
      I: ProvisionRequest,
      O: ProvisionResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc xyz.block.ftl.v1beta1.provisioner.ProvisionerPluginService.Plan
     */
    plan: {
      name: "Plan",
      I: PlanRequest,
      O: PlanResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc xyz.block.ftl.v1beta1.provisioner.ProvisionerPluginService.Status
     */
    status: {
      name: "Status",
      I: StatusRequest,
      O: StatusResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

