// @generated by protoc-gen-connect-es v1.6.1 with parameter "target=ts"
// @generated from file xyz/block/ftl/language/v1/language.proto (package xyz.block.ftl.language.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "../../v1/ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { BuildContextUpdatedRequest, BuildContextUpdatedResponse, BuildRequest, BuildResponse, CreateModuleRequest, CreateModuleResponse, GenerateStubsRequest, GenerateStubsResponse, GetCreateModuleFlagsRequest, GetCreateModuleFlagsResponse, GetDependenciesRequest, GetDependenciesResponse, ModuleConfigDefaultsRequest, ModuleConfigDefaultsResponse, SyncStubReferencesRequest, SyncStubReferencesResponse } from "./language_pb.js";

/**
 * LanguageService allows a plugin to add support for a programming language.
 *
 * @generated from service xyz.block.ftl.language.v1.LanguageService
 */
export const LanguageService = {
  typeName: "xyz.block.ftl.language.v1.LanguageService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Get language specific flags that can be used to create a new module.
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.GetCreateModuleFlags
     */
    getCreateModuleFlags: {
      name: "GetCreateModuleFlags",
      I: GetCreateModuleFlagsRequest,
      O: GetCreateModuleFlagsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Generates files for a new module with the requested name
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.CreateModule
     */
    createModule: {
      name: "CreateModule",
      I: CreateModuleRequest,
      O: CreateModuleResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.ModuleConfigDefaults
     */
    moduleConfigDefaults: {
      name: "ModuleConfigDefaults",
      I: ModuleConfigDefaultsRequest,
      O: ModuleConfigDefaultsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Extract dependencies for a module
     * FTL will ensure that these dependencies are built before requesting a build for this module.
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.GetDependencies
     */
    getDependencies: {
      name: "GetDependencies",
      I: GetDependenciesRequest,
      O: GetDependenciesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Build the module and stream back build events.
     *
     * A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
     * end of the build.
     *
     * The request can include the option to "rebuild_automatically". In this case the plugin should watch for
     * file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
     * rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
     * calls.
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.Build
     */
    build: {
      name: "Build",
      I: BuildRequest,
      O: BuildResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
     * build context is updated.
     *
     * Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
     * event with the updated build context id with "is_automatic_rebuild" as false.
     *
     * If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
     * build stream, it must fail the BuildContextUpdated call.
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.BuildContextUpdated
     */
    buildContextUpdated: {
      name: "BuildContextUpdated",
      I: BuildContextUpdatedRequest,
      O: BuildContextUpdatedResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Generate stubs for a module.
     *
     * Stubs allow modules to import other module's exported interface. If a language does not need this step,
     * then it is not required to do anything in this call.
     *
     * This call is not tied to the module that this plugin is responsible for. A plugin of each language will
     * be chosen to generate stubs for each module.
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.GenerateStubs
     */
    generateStubs: {
      name: "GenerateStubs",
      I: GenerateStubsRequest,
      O: GenerateStubsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
     * references to external modules, regardless of whether they are dependencies.
     *
     * For example, go plugin adds references to all modules into the go.work file so that tools can automatically
     * import the modules when users start reference them.
     *
     * It is optional to do anything with this call.
     *
     * @generated from rpc xyz.block.ftl.language.v1.LanguageService.SyncStubReferences
     */
    syncStubReferences: {
      name: "SyncStubReferences",
      I: SyncStubReferencesRequest,
      O: SyncStubReferencesResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

