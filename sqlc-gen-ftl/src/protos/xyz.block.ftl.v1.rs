// @generated
// This file is @generated by prost-build.
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct PingRequest {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PingResponse {
    /// If present, the service is not ready to accept requests and this is the
    /// reason.
    #[prost(string, optional, tag="1")]
    pub not_ready: ::core::option::Option<::prost::alloc::string::String>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Metadata {
    #[prost(message, repeated, tag="1")]
    pub values: ::prost::alloc::vec::Vec<metadata::Pair>,
}
/// Nested message and enum types in `Metadata`.
pub mod metadata {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Pair {
        #[prost(string, tag="1")]
        pub key: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub value: ::prost::alloc::string::String,
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RegisterRunnerRequest {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub endpoint: ::prost::alloc::string::String,
    #[prost(string, tag="3")]
    pub deployment: ::prost::alloc::string::String,
    #[prost(message, optional, tag="5")]
    pub labels: ::core::option::Option<::prost_types::Struct>,
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct RegisterRunnerResponse {
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct StatusRequest {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StatusResponse {
    #[prost(message, repeated, tag="1")]
    pub controllers: ::prost::alloc::vec::Vec<status_response::Controller>,
    #[prost(message, repeated, tag="2")]
    pub runners: ::prost::alloc::vec::Vec<status_response::Runner>,
    #[prost(message, repeated, tag="3")]
    pub deployments: ::prost::alloc::vec::Vec<status_response::Deployment>,
    #[prost(message, repeated, tag="5")]
    pub routes: ::prost::alloc::vec::Vec<status_response::Route>,
}
/// Nested message and enum types in `StatusResponse`.
pub mod status_response {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Controller {
        #[prost(string, tag="1")]
        pub key: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub endpoint: ::prost::alloc::string::String,
        #[prost(string, tag="3")]
        pub version: ::prost::alloc::string::String,
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Runner {
        #[prost(string, tag="1")]
        pub key: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub endpoint: ::prost::alloc::string::String,
        #[prost(string, optional, tag="3")]
        pub deployment: ::core::option::Option<::prost::alloc::string::String>,
        #[prost(message, optional, tag="4")]
        pub labels: ::core::option::Option<::prost_types::Struct>,
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Deployment {
        #[prost(string, tag="1")]
        pub key: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub language: ::prost::alloc::string::String,
        #[prost(string, tag="3")]
        pub name: ::prost::alloc::string::String,
        #[prost(int32, tag="4")]
        pub min_replicas: i32,
        #[prost(int32, tag="7")]
        pub replicas: i32,
        #[prost(message, optional, tag="5")]
        pub labels: ::core::option::Option<::prost_types::Struct>,
        #[prost(message, optional, tag="6")]
        pub schema: ::core::option::Option<super::super::schema::v1::Module>,
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Route {
        #[prost(string, tag="1")]
        pub module: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub deployment: ::prost::alloc::string::String,
        #[prost(string, tag="3")]
        pub endpoint: ::prost::alloc::string::String,
    }
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct ProcessListRequest {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProcessListResponse {
    #[prost(message, repeated, tag="1")]
    pub processes: ::prost::alloc::vec::Vec<process_list_response::Process>,
}
/// Nested message and enum types in `ProcessListResponse`.
pub mod process_list_response {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct ProcessRunner {
        #[prost(string, tag="1")]
        pub key: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub endpoint: ::prost::alloc::string::String,
        #[prost(message, optional, tag="3")]
        pub labels: ::core::option::Option<::prost_types::Struct>,
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Process {
        #[prost(string, tag="1")]
        pub deployment: ::prost::alloc::string::String,
        #[prost(int32, tag="2")]
        pub min_replicas: i32,
        #[prost(message, optional, tag="3")]
        pub labels: ::core::option::Option<::prost_types::Struct>,
        #[prost(message, optional, tag="4")]
        pub runner: ::core::option::Option<ProcessRunner>,
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDeploymentContextRequest {
    #[prost(string, tag="1")]
    pub deployment: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDeploymentContextResponse {
    #[prost(string, tag="1")]
    pub module: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub deployment: ::prost::alloc::string::String,
    #[prost(map="string, bytes", tag="3")]
    pub configs: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::bytes::Bytes>,
    #[prost(map="string, bytes", tag="4")]
    pub secrets: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::bytes::Bytes>,
    #[prost(message, repeated, tag="5")]
    pub databases: ::prost::alloc::vec::Vec<get_deployment_context_response::Dsn>,
    #[prost(message, repeated, tag="6")]
    pub routes: ::prost::alloc::vec::Vec<get_deployment_context_response::Route>,
    #[prost(map="string, string", tag="7")]
    pub egress: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
/// Nested message and enum types in `GetDeploymentContextResponse`.
pub mod get_deployment_context_response {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Dsn {
        #[prost(string, tag="1")]
        pub name: ::prost::alloc::string::String,
        #[prost(enumeration="DbType", tag="2")]
        pub r#type: i32,
        #[prost(string, tag="3")]
        pub dsn: ::prost::alloc::string::String,
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Route {
        #[prost(string, tag="1")]
        pub deployment: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub uri: ::prost::alloc::string::String,
    }
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum DbType {
        Unspecified = 0,
        Postgres = 1,
        Mysql = 2,
    }
    impl DbType {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Self::Unspecified => "DB_TYPE_UNSPECIFIED",
                Self::Postgres => "DB_TYPE_POSTGRES",
                Self::Mysql => "DB_TYPE_MYSQL",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "DB_TYPE_UNSPECIFIED" => Some(Self::Unspecified),
                "DB_TYPE_POSTGRES" => Some(Self::Postgres),
                "DB_TYPE_MYSQL" => Some(Self::Mysql),
                _ => None,
            }
        }
    }
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct GetSchemaRequest {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetSchemaResponse {
    #[prost(message, optional, tag="1")]
    pub schema: ::core::option::Option<super::schema::v1::Schema>,
    #[prost(message, repeated, tag="2")]
    pub changesets: ::prost::alloc::vec::Vec<super::schema::v1::Changeset>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PullSchemaRequest {
    #[prost(string, tag="1")]
    pub subscription_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PullSchemaResponse {
    #[prost(message, optional, tag="1")]
    pub event: ::core::option::Option<super::schema::v1::Notification>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UpdateDeploymentRuntimeRequest {
    #[prost(string, optional, tag="1")]
    pub changeset: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(message, optional, tag="2")]
    pub update: ::core::option::Option<super::schema::v1::RuntimeElement>,
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct UpdateDeploymentRuntimeResponse {
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct GetDeploymentsRequest {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDeploymentsResponse {
    #[prost(message, repeated, tag="1")]
    pub schema: ::prost::alloc::vec::Vec<DeployedSchema>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RealmChange {
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    /// The modules to add or update.
    #[prost(message, repeated, tag="2")]
    pub modules: ::prost::alloc::vec::Vec<super::schema::v1::Module>,
    /// The deployments to remove.
    #[prost(string, repeated, tag="3")]
    pub to_remove: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// Whether this is an external realm.
    #[prost(bool, tag="4")]
    pub external: bool,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateChangesetRequest {
    #[prost(message, repeated, tag="1")]
    pub realm_changes: ::prost::alloc::vec::Vec<RealmChange>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateChangesetResponse {
    /// The changeset key of the newly created changeset.
    #[prost(string, tag="1")]
    pub changeset: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeployedSchema {
    #[prost(string, tag="1")]
    pub deployment_key: ::prost::alloc::string::String,
    #[prost(message, optional, tag="2")]
    pub schema: ::core::option::Option<super::schema::v1::Module>,
    #[prost(bool, tag="3")]
    pub is_active: bool,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PrepareChangesetRequest {
    /// The changeset key to prepare.
    #[prost(string, tag="1")]
    pub changeset: ::prost::alloc::string::String,
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct PrepareChangesetResponse {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CommitChangesetRequest {
    /// The changeset key to commit.
    #[prost(string, tag="1")]
    pub changeset: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CommitChangesetResponse {
    #[prost(message, optional, tag="1")]
    pub changeset: ::core::option::Option<super::schema::v1::Changeset>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DrainChangesetRequest {
    /// The changeset key to commit.
    #[prost(string, tag="1")]
    pub changeset: ::prost::alloc::string::String,
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct DrainChangesetResponse {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FinalizeChangesetRequest {
    /// The changeset key to commit.
    #[prost(string, tag="1")]
    pub changeset: ::prost::alloc::string::String,
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct FinalizeChangesetResponse {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FailChangesetRequest {
    /// The changeset key to fail.
    #[prost(string, tag="1")]
    pub changeset: ::prost::alloc::string::String,
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct FailChangesetResponse {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RollbackChangesetRequest {
    /// The changeset key to fail.
    #[prost(string, tag="1")]
    pub changeset: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub error: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RollbackChangesetResponse {
    #[prost(message, optional, tag="1")]
    pub changeset: ::core::option::Option<super::schema::v1::Changeset>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDeploymentRequest {
    #[prost(string, tag="1")]
    pub deployment_key: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetDeploymentResponse {
    #[prost(message, optional, tag="1")]
    pub schema: ::core::option::Option<super::schema::v1::Module>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CallRequest {
    #[prost(message, optional, tag="1")]
    pub metadata: ::core::option::Option<Metadata>,
    #[prost(message, optional, tag="2")]
    pub verb: ::core::option::Option<super::schema::v1::Ref>,
    #[prost(bytes="bytes", tag="3")]
    pub body: ::prost::bytes::Bytes,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CallResponse {
    #[prost(oneof="call_response::Response", tags="1, 2")]
    pub response: ::core::option::Option<call_response::Response>,
}
/// Nested message and enum types in `CallResponse`.
pub mod call_response {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Error {
        #[prost(string, tag="1")]
        pub message: ::prost::alloc::string::String,
        /// TODO: Richer error type.
        #[prost(string, optional, tag="2")]
        pub stack: ::core::option::Option<::prost::alloc::string::String>,
    }
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Response {
        #[prost(bytes, tag="1")]
        Body(::prost::bytes::Bytes),
        #[prost(message, tag="2")]
        Error(Error),
    }
}
// @@protoc_insertion_point(module)
