use std::process::Command;
use std::path::PathBuf;
use std::fs;
use prost::Message;
use tempfile::TempDir;
use sha2::{Sha256, Digest};

#[path = "../src/plugin/mod.rs"]
mod plugin;
#[path = "../src/protos/mod.rs"]
mod protos;

use protos::schemapb;

fn build_wasm() -> Result<(), Box<dyn std::error::Error>> {
    let status = Command::new("just")
        .arg("build-sqlc-gen-ftl")
        .status()?;

    if !status.success() {
        return Err("Failed to build WASM".into());
    }
    Ok(())
}

fn get_test_queries(engine: &str) -> Vec<String> {
    match engine {
        "mysql" => vec![
            "SELECT id, big_int, small_int, some_decimal, some_numeric, some_float, some_double, some_varchar, some_text, some_char, nullable_text, some_bool, nullable_bool, some_date, some_time, some_timestamp, some_blob, some_json FROM all_types WHERE id = ?".to_string(),
            "INSERT INTO all_types (     big_int, small_int,      some_decimal, some_numeric, some_float, some_double,     some_varchar, some_text, some_char, nullable_text,     some_bool, nullable_bool,     some_date, some_time, some_timestamp,     some_blob, some_json ) VALUES (     ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )".to_string(),
        ],
        "postgresql" => vec![
            "SELECT id, big_int, small_int, some_decimal, some_numeric, some_float, some_double, some_varchar, some_text, some_char, nullable_text, some_bool, nullable_bool, some_date, some_time, some_timestamp, some_blob, some_json FROM all_types WHERE id = $1".to_string(),
            "INSERT INTO all_types (     big_int, small_int,      some_decimal, some_numeric, some_float, some_double,     some_varchar, some_text, some_char, nullable_text,     some_bool, nullable_bool,     some_date, some_time, some_timestamp,     some_blob, some_json ) VALUES (     $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17 )".to_string(),
        ],
        _ => vec![],
    }
}

fn get_sqlc_config(wasm_path: &PathBuf, engine: &str) -> Result<String, Box<dyn std::error::Error>> {
    let wasm_contents = fs::read(wasm_path)?;
    let mut hasher = Sha256::new();
    hasher.update(&wasm_contents);
    let sha256_hash = hex::encode(hasher.finalize());

    Ok(format!(
        r#"version: '2'
plugins:
- name: ftl
  wasm:
    url: file://sqlc-gen-ftl.wasm
    sha256: {}
sql:
- schema: schema.sql
  queries: queries.sql
  engine: {}
  codegen:
  - out: gen
    plugin: ftl
    options:
      module: echo"#,
      sha256_hash,
      engine,
    ))
}

fn expected_module_schema(engine: &str) -> schemapb::Module {
    let queries = get_test_queries(engine);
    let fields = vec![
        ("bigInt", schemapb::r#type::Value::Int(schemapb::Int { pos: None }), "big_int"),
        ("smallInt", schemapb::r#type::Value::Int(schemapb::Int { pos: None }), "small_int"),
        ("someDecimal", schemapb::r#type::Value::String(schemapb::String { pos: None }), "some_decimal"),
        ("someNumeric", schemapb::r#type::Value::String(schemapb::String { pos: None }), "some_numeric"),
        ("someFloat", schemapb::r#type::Value::Float(schemapb::Float { pos: None }), "some_float"),
        ("someDouble", schemapb::r#type::Value::Float(schemapb::Float { pos: None }), "some_double"),
        ("someVarchar", schemapb::r#type::Value::String(schemapb::String { pos: None }), "some_varchar"),
        ("someText", schemapb::r#type::Value::String(schemapb::String { pos: None }), "some_text"),
        ("someChar", schemapb::r#type::Value::String(schemapb::String { pos: None }), "some_char"),
        ("nullableText", schemapb::r#type::Value::Optional(Box::new(schemapb::Optional {
            pos: None,
            r#type: Some(Box::new(schemapb::Type {
                value: Some(schemapb::r#type::Value::String(schemapb::String { pos: None }))
            }))
        })), "nullable_text"),
        ("someBool", schemapb::r#type::Value::Bool(schemapb::Bool { pos: None }), "some_bool"),
        ("nullableBool", schemapb::r#type::Value::Optional(Box::new(schemapb::Optional {
            pos: None,
            r#type: Some(Box::new(schemapb::Type {
                value: Some(schemapb::r#type::Value::Bool(schemapb::Bool { pos: None }))
            }))
        })), "nullable_bool"),
        ("someDate", schemapb::r#type::Value::Time(schemapb::Time { pos: None }), "some_date"),
        ("someTime", schemapb::r#type::Value::Time(schemapb::Time { pos: None }), "some_time"),
        ("someTimestamp", schemapb::r#type::Value::Time(schemapb::Time { pos: None }), "some_timestamp"),
        ("someBlob", schemapb::r#type::Value::Bytes(schemapb::Bytes { pos: None }), "some_blob"),
        ("someJson", schemapb::r#type::Value::Any(schemapb::Any { pos: None }), "some_json"),
    ];

    let create_field = |name: &str, type_value: schemapb::r#type::Value, db_name: &str| {
        schemapb::Field {
            name: name.to_string(),
            r#type: Some(schemapb::Type {
                value: Some(type_value)
            }),
            pos: None,
            comments: vec![],
            metadata: vec![schemapb::Metadata {
                value: Some(schemapb::metadata::Value::DbColumn(schemapb::MetadataDbColumn {
                    pos: None,
                    table: "all_types".to_string(),
                    name: db_name.to_string(),
                }))
            }],
        }
    };

    let create_fields = |fields: &[(& str, schemapb::r#type::Value, &str)]| {
        fields.iter().map(|(name, type_value, db_name)| {
            create_field(name, type_value.clone(), db_name)
        }).collect::<Vec<_>>()
    };

    schemapb::Module {
        name: "echo".to_string(),
        builtin: false,
        runtime: None,
        comments: vec![],
        metadata: vec![],
        pos: None,
        decls: vec![
            // CreateAllTypesQuery
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Data(schemapb::Data {
                    name: "CreateAllTypesQuery".to_string(),
                    export: false,
                    type_parameters: vec![],
                    fields: create_fields(&fields),
                    pos: None,
                    comments: vec![],
                    metadata: vec![],
                })),
            },
            // GetAllTypesQuery
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Data(schemapb::Data {
                    name: "GetAllTypesQuery".to_string(),
                    export: false,
                    type_parameters: vec![],
                    fields: vec![
                        schemapb::Field {
                            pos: None,
                            comments: vec![],
                            name: "id".to_string(),
                            r#type: Some(schemapb::Type {
                                value: Some(schemapb::r#type::Value::Int(schemapb::Int { pos: None })),
                            }),
                            metadata: vec![schemapb::Metadata {
                                value: Some(schemapb::metadata::Value::DbColumn(schemapb::MetadataDbColumn {
                                    pos: None,
                                    table: "all_types".to_string(),
                                    name: "id".to_string(),
                                })),
                            }],
                        },
                    ],
                    pos: None,
                    comments: vec![],
                    metadata: vec![],
                })),
            },
            // GetAllTypesResult
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Data(schemapb::Data {
                    name: "GetAllTypesResult".to_string(),
                    export: false,
                    type_parameters: vec![],
                    fields: {
                        let mut result_fields = vec![
                            schemapb::Field {
                                name: "id".to_string(),
                                r#type: Some(schemapb::Type {
                                    value: Some(schemapb::r#type::Value::Int(schemapb::Int { pos: None }))
                                }),
                                pos: None,
                                comments: vec![],
                                metadata: vec![schemapb::Metadata {
                                    value: Some(schemapb::metadata::Value::DbColumn(schemapb::MetadataDbColumn {
                                        pos: None,
                                        table: "all_types".to_string(),
                                        name: "id".to_string(),
                                    }))
                                }],
                            },
                        ];
                        result_fields.extend(create_fields(&fields));
                        result_fields
                    },
                    pos: None,
                    comments: vec![],
                    metadata: vec![],
                })),
            },
            // CreateAllTypes verb
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Verb(schemapb::Verb {
                    name: "createAllTypes".to_string(),
                    export: false,
                    runtime: None,
                    request: Some(schemapb::Type {
                        value: Some(schemapb::r#type::Value::Ref(schemapb::Ref {
                            module: "echo".to_string(),
                            name: "CreateAllTypesQuery".to_string(),
                            pos: None,
                            type_parameters: vec![],
                        }))
                    }),
                    response: Some(schemapb::Type {
                        value: Some(schemapb::r#type::Value::Unit(schemapb::Unit { pos: None }))
                    }),
                    pos: None,
                    comments: vec![],
                    metadata: vec![schemapb::Metadata {
                        value: Some(schemapb::metadata::Value::SqlQuery(schemapb::MetadataSqlQuery {
                            pos: None,
                            command: "exec".to_string(),
                            query: queries[1].clone(),
                        })),
                    }],
                })),
            },
            // GetAllTypes verb
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Verb(schemapb::Verb {
                    name: "getAllTypes".to_string(),
                    export: false,
                    runtime: None,
                    request: Some(schemapb::Type {
                        value: Some(schemapb::r#type::Value::Ref(schemapb::Ref {
                            module: "echo".to_string(),
                            name: "GetAllTypesQuery".to_string(),
                            pos: None,
                            type_parameters: vec![],
                        }))
                    }),
                    response: Some(schemapb::Type {
                        value: Some(schemapb::r#type::Value::Ref(schemapb::Ref {
                            module: "echo".to_string(),
                            name: "GetAllTypesResult".to_string(),
                            pos: None,
                            type_parameters: vec![],
                        }))
                    }),
                    pos: None,
                    comments: vec![],
                    metadata: vec![schemapb::Metadata {
                        value: Some(schemapb::metadata::Value::SqlQuery(schemapb::MetadataSqlQuery {
                            pos: None,
                            command: "one".to_string(),
                            query: queries[0].clone(),
                        })),
                    }],
                })),
            },
        ],
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    #[cfg_attr(not(feature = "ci"), ignore)]
    fn test_wasm_generate() -> Result<(), Box<dyn std::error::Error>> {
        if let Err(e) = build_wasm() {
            return Err(format!("Failed to build WASM: {}", e).into());
        }

        for engine in ["mysql", "postgresql"] {
            let temp_dir = TempDir::new()?;
            let gen_dir = temp_dir.path().join("gen");
            std::fs::create_dir(&gen_dir)?;
            
            let root_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
            let test_dir = root_dir.join("test");
            let wasm_path = temp_dir.path().join("sqlc-gen-ftl.wasm");

            std::fs::copy(
                test_dir.join(format!("testdata/{}/schema.sql", engine)),
                temp_dir.path().join("schema.sql")
            )?;
            std::fs::copy(
                test_dir.join(format!("testdata/{}/queries.sql", engine)),
                temp_dir.path().join("queries.sql")
            )?;
            std::fs::copy(
                root_dir.join("../internal/sqlc/resources/sqlc-gen-ftl.wasm"),
                &wasm_path
            )?;
            
            let config_contents = get_sqlc_config(&wasm_path, engine)?;
            let config_path = temp_dir.path().join("sqlc.yaml");
            std::fs::write(&config_path, config_contents)?;

            let output = Command::new("sqlc")
                .arg("generate")
                .arg("--file")
                .arg(&config_path)
                .current_dir(temp_dir.path())
                .output()?;

            if !output.status.success() {
                return Err(format!(
                    "sqlc generate failed for {} with status: {}\nstderr: {}",
                    engine,
                    output.status,
                    String::from_utf8_lossy(&output.stderr)
                ).into());
            }

            let pb_contents = std::fs::read(gen_dir.join("queries.pb"))?;
            let mut actual_module = schemapb::Module::decode(&*pb_contents)?;
            let mut expected_module = expected_module_schema(engine);

            // Sort declarations by name before comparison
            actual_module.decls.sort_by(|a, b| {
                let name_a = match &a.value {
                    Some(schemapb::decl::Value::Data(d)) => &d.name,
                    Some(schemapb::decl::Value::Verb(v)) => &v.name,
                    _ => "",
                };
                let name_b = match &b.value {
                    Some(schemapb::decl::Value::Data(d)) => &d.name,
                    Some(schemapb::decl::Value::Verb(v)) => &v.name,
                    _ => "",
                };
                name_a.cmp(name_b)
            });
            expected_module.decls.sort_by(|a, b| {
                let name_a = match &a.value {
                    Some(schemapb::decl::Value::Data(d)) => &d.name,
                    Some(schemapb::decl::Value::Verb(v)) => &v.name,
                    _ => "",
                };
                let name_b = match &b.value {
                    Some(schemapb::decl::Value::Data(d)) => &d.name,
                    Some(schemapb::decl::Value::Verb(v)) => &v.name,
                    _ => "",
                };
                name_a.cmp(name_b)
            });

            assert_eq!(
                &actual_module, 
                &expected_module, 
                "Schema mismatch for {}.\nActual: {:#?}\nExpected: {:#?}",
                engine,
                actual_module,
                expected_module
            );
        }

        Ok(())
    }
}

