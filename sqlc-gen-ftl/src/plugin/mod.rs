#![allow(dead_code)]

use crate::protos::pluginpb;
use crate::protos::schemapb;
use crate::protos::schemapb::r#type::Value as TypeValue;
use convert_case::{Case, Casing};
use prost::Message;
use std::io;

pub struct Plugin;

impl Plugin {
    pub fn generate_from_input(input: &[u8]) -> Result<Vec<u8>, io::Error> {
        let req = pluginpb::GenerateRequest::decode(input)
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))?;
        let resp = Self::handle_generate(req)?;
        Ok(resp.encode_to_vec())
    }

    fn handle_generate(req: pluginpb::GenerateRequest) -> Result<pluginpb::GenerateResponse, io::Error> {
        let module = generate_schema(&req)?;
        Ok(pluginpb::GenerateResponse {
            files: vec![pluginpb::File {
                name: "queries.pb".to_string(),
                contents: module.encode_to_vec(),
            }],
        })
    }
}

fn generate_schema(request: &pluginpb::GenerateRequest) -> Result<schemapb::Module, io::Error> {
    let mut decls = Vec::new();
    let module_name = get_module_name(request)?;
    
    for query in &request.queries {
        if !query.params.is_empty() {
            decls.push(to_verb_request(query, request));
        }

        if !query.columns.is_empty() {
            decls.push(to_verb_response(query, request));
        }

        decls.push(to_verb(query, &module_name));
    }

    Ok(schemapb::Module {
        name: module_name,
        builtin: false,
        runtime: None,
        comments: Vec::new(),
        metadata: Vec::new(),
        pos: None,
        decls,
    })
}

fn to_verb(query: &pluginpb::Query, module_name: &str) -> schemapb::Decl {
    let upper_camel_name = to_upper_camel(&query.name);
    let request_type = if !query.params.is_empty() {
        Some(to_schema_ref(module_name, &format!("{}Query", upper_camel_name)))
    } else {
        Some(to_schema_unit())
    };

    let response_type = match query.cmd.as_str() {
        ":exec" => Some(to_schema_unit()),
        ":one" => Some(to_schema_ref(module_name, &format!("{}Result", upper_camel_name))),
        ":many" => Some(schemapb::Type {
            value: Some(schemapb::r#type::Value::Array(Box::new(schemapb::Array {
                pos: None,
                element: Some(Box::new(to_schema_ref(module_name, &format!("{}Result", upper_camel_name)))),
            }))),
        }),
        _ => Some(to_schema_unit()),
    };

    let sql_query_metadata = schemapb::Metadata {
        value: Some(schemapb::metadata::Value::SqlQuery(schemapb::MetadataSqlQuery {
            pos: None,
            query: query.text.replace('\n', " ").trim().to_string(),
            command: query.cmd.trim_start_matches(':').to_string(),
        })),
    };

    schemapb::Decl {
        value: Some(schemapb::decl::Value::Verb(schemapb::Verb {
            name: query.name.to_case(Case::Camel),
            export: false,
            runtime: None,
            request: request_type,
            response: response_type,
            pos: None,
            comments: Vec::new(),
            metadata: vec![sql_query_metadata],
        })),
    }
}

fn to_verb_request(query: &pluginpb::Query, req: &pluginpb::GenerateRequest) -> schemapb::Decl {
    let upper_camel_name = to_upper_camel(&query.name);
    schemapb::Decl {
        value: Some(schemapb::decl::Value::Data(schemapb::Data {
            name: format!("{}Query", upper_camel_name),
            export: false,
            type_parameters: Vec::new(),
            fields: query.params.iter().map(|param| {
                let name = if let Some(col) = &param.column {
                    col.name.clone()
                } else {
                    format!("arg{}", param.number)
                };
                
                let sql_type = param.column.as_ref().and_then(|col| col.r#type.as_ref());
                to_schema_field(name, param.column.as_ref(), sql_type, req)
            }).collect(),
            pos: None,
            comments: Vec::new(),
            metadata: Vec::new(),
        })),
    }
}

fn to_verb_response(query: &pluginpb::Query, req: &pluginpb::GenerateRequest) -> schemapb::Decl {
    let pascal_name = to_upper_camel(&query.name);
    schemapb::Decl {
        value: Some(schemapb::decl::Value::Data(schemapb::Data {
            name: format!("{}Result", pascal_name),
            export: false,
            type_parameters: Vec::new(),
            fields: query.columns.iter().map(|col| {
                to_schema_field(col.name.clone(), Some(col), col.r#type.as_ref(), req)
            }).collect(),
            pos: None,
            comments: Vec::new(),
            metadata: Vec::new(),
        })),
    }
}

fn to_schema_field(name: String, col: Option<&pluginpb::Column>, sql_type: Option<&pluginpb::Identifier>, req: &pluginpb::GenerateRequest) -> schemapb::Field {
    let mut metadata = Vec::new();

    if let Some(col) = col {
        if let Some(table) = &col.table {
            let db_column = schemapb::Metadata {
                value: Some(schemapb::metadata::Value::DbColumn(schemapb::MetadataDbColumn {
                    pos: None,
                    table: table.name.clone(),
                    name: col.name.clone(),
                })),
            };
            metadata.push(db_column);
        }
    }

    schemapb::Field {
        name: name.to_case(Case::Camel),
        r#type: Some(match (col, sql_type) {
            (Some(col), _) => to_schema_type(req, col),
            (None, _) => schemapb::Type {
                value: Some(schemapb::r#type::Value::Any(schemapb::Any { pos: None })),
            },
        }),
        pos: None,
        comments: Vec::new(),
        metadata,
    }
}

fn to_schema_ref(module_name: &str, name: &str) -> schemapb::Type {
    schemapb::Type {
        value: Some(schemapb::r#type::Value::Ref(schemapb::Ref {
            module: module_name.to_string(),
            name: name.to_string(),
            pos: None,
            type_parameters: vec![],
        }))
    }
}

fn to_schema_unit() -> schemapb::Type {
    schemapb::Type {
        value: Some(schemapb::r#type::Value::Unit(schemapb::Unit {
            pos: None,
        }))
    }
}

fn get_module_name(req: &pluginpb::GenerateRequest) -> Result<String, io::Error> {
    let codegen = req.settings
        .as_ref()
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidData, "Missing settings"))?
        .codegen
        .as_ref()
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidData, "Missing codegen settings"))?;

    let options_str = String::from_utf8(codegen.options.clone())
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, format!("Invalid UTF-8 in options: {}", e)))?;
    
    let options: serde_json::Value = serde_json::from_str(&options_str)
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, format!("Failed to parse JSON options: {}", e)))?;

    options.get("module")
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidData, "Missing module name in options"))
}

fn to_upper_camel(s: &str) -> String {
    let snake = s.to_case(Case::Snake);
    snake.split('_')
        .map(|part| part.to_case(Case::Title))
        .collect()
}

fn to_schema_type(req: &pluginpb::GenerateRequest, col: &pluginpb::Column) -> schemapb::Type {
    let engine = req.settings
        .as_ref()
        .and_then(|s| Some(s.engine.as_str()))
        .unwrap_or("mysql");

    match engine {
        "mysql" => mysql_to_schema_type(col),
        "postgresql" | "postgres" => postgresql_to_schema_type(col),
        _ => mysql_to_schema_type(col)
    }
}

fn mysql_to_schema_type(col: &pluginpb::Column) -> schemapb::Type {
    let column_type = col.r#type.as_ref()
        .map(|t| t.name.to_lowercase())
        .unwrap_or_default();
    let not_null = col.not_null || col.is_array;
    let length = col.length;

    let value = match column_type.as_str() {
        "varchar" | "text" | "char" | "tinytext" | "mediumtext" | "longtext" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None }))
                    }))
                }))
            }
        },
        "tinyint" => {
            if length == 1 {
                if not_null {
                    TypeValue::Bool(schemapb::Bool { pos: None })
                } else {
                    TypeValue::Optional(Box::new(schemapb::Optional {
                        pos: None,
                        r#type: Some(Box::new(schemapb::Type {
                            value: Some(TypeValue::Bool(schemapb::Bool { pos: None }))
                        }))
                    }))
                }
            } else {
                if not_null {
                    TypeValue::Int(schemapb::Int { pos: None })
                } else {
                    TypeValue::Optional(Box::new(schemapb::Optional {
                        pos: None,
                        r#type: Some(Box::new(schemapb::Type {
                            value: Some(TypeValue::Int(schemapb::Int { pos: None }))
                        }))
                    }))
                }
            }
        },
        "year" | "smallint" => {
            if not_null {
                TypeValue::Int(schemapb::Int { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Int(schemapb::Int { pos: None }))
                    }))
                }))
            }
        },
        "int" | "integer" | "mediumint" | "bigint" => {
            if not_null {
                TypeValue::Int(schemapb::Int { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Int(schemapb::Int { pos: None }))
                    }))
                }))
            }
        },
        "blob" | "binary" | "varbinary" | "tinyblob" | "mediumblob" | "longblob" => {
            if not_null {
                TypeValue::Bytes(schemapb::Bytes { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Bytes(schemapb::Bytes { pos: None }))
                    }))
                }))
            }
        },
        "double" | "double precision" | "real" | "float" => {
            if not_null {
                TypeValue::Float(schemapb::Float { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Float(schemapb::Float { pos: None }))
                    }))
                }))
            }
        },
        "decimal" | "dec" | "fixed" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None }))
                    }))
                }))
            }
        },
        "enum" => TypeValue::String(schemapb::String { pos: None }),
        "date" | "timestamp" | "datetime" | "time" => {
            if not_null {
                TypeValue::Time(schemapb::Time { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Time(schemapb::Time { pos: None }))
                    }))
                }))
            }
        },
        "boolean" | "bool" => {
            if not_null {
                TypeValue::Bool(schemapb::Bool { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Bool(schemapb::Bool { pos: None }))
                    }))
                }))
            }
        },
        "json" => TypeValue::Any(schemapb::Any { pos: None }),
        "any" => TypeValue::Any(schemapb::Any { pos: None }),
        _ => {
            // TODO: Handle enum types from catalog
            TypeValue::Any(schemapb::Any { pos: None })
        }
    };

    schemapb::Type {
        value: Some(value),
    }
}

fn postgresql_to_schema_type(col: &pluginpb::Column) -> schemapb::Type {
    let column_type = col.r#type.as_ref()
        .map(|t| t.name.to_lowercase())
        .unwrap_or_default();
    let not_null = col.not_null || col.is_array;

    let value = match column_type.as_str() {
        "smallint" | "int2" | "pg_catalog.int2" |
        "integer" | "int" | "int4" | "pg_catalog.int4" |
        "bigint" | "int8" | "pg_catalog.int8" |
        "smallserial" | "serial2" | "pg_catalog.serial2" |
        "serial" | "serial4" | "pg_catalog.serial4" |
        "bigserial" | "serial8" | "pg_catalog.serial8"  => {
            if not_null {
                TypeValue::Int(schemapb::Int { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Int(schemapb::Int { pos: None }))
                    }))
                }))
            }
        },

        "float" | "double precision" | "float8" | "pg_catalog.float8" |
        "real" | "float4" | "pg_catalog.float4" => {
            if not_null {
                TypeValue::Float(schemapb::Float { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Float(schemapb::Float { pos: None }))
                    }))
                }))
            }
        },
        
        "numeric" | "decimal" | "pg_catalog.numeric" | "money" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None }))
                    }))
                }))
            }
        },

        "boolean" | "bool" | "pg_catalog.bool" => {
            if not_null {
                TypeValue::Bool(schemapb::Bool { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Bool(schemapb::Bool { pos: None }))
                    }))
                }))
            }
        },

        "json" | "jsonb" => TypeValue::Any(schemapb::Any { pos: None }),

        "bytea" | "blob" | "pg_catalog.bytea" => TypeValue::Bytes(schemapb::Bytes { pos: None }),

        "date" | "pg_catalog.time" | "pg_catalog.timetz" | 
        "pg_catalog.timestamp" | "pg_catalog.timestamptz" | "timestamptz" |
        "time" | "timestamp" => {
            if not_null {
                TypeValue::Time(schemapb::Time { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Time(schemapb::Time { pos: None }))
                    }))
                }))
            }
        },

        "text" | "pg_catalog.varchar" | "pg_catalog.bpchar" | "string" | 
        "citext" | "name" | "uuid" | "ltree" | "lquery" | "ltxtquery" |
        "varchar" | "char" | "character" | "pg_catalog.char" | "pg_catalog.character" | "bpchar" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None }))
                    }))
                }))
            }
        },

        "inet" | "cidr" | "macaddr" | "macaddr8" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None }))
                    }))
                }))
            }
        },

        "box" | "circle" | "line" | "lseg" | "path" | "point" | "polygon" => {
            TypeValue::String(schemapb::String { pos: None })
        },

        "bit" | "varbit" | "pg_catalog.bit" | "pg_catalog.varbit" => {
            TypeValue::String(schemapb::String { pos: None })
        },

        "daterange" | "tsrange" | "tstzrange" | "numrange" | 
        "int4range" | "int8range" | 
        "datemultirange" | "tsmultirange" | "tstzmultirange" | 
        "nummultirange" | "int4multirange" | "int8multirange" => {
            TypeValue::Any(schemapb::Any { pos: None })
        },

        "void" | "any" => TypeValue::Any(schemapb::Any { pos: None }),
        _ => {
            // TODO: Handle enum types from catalog
            TypeValue::Any(schemapb::Any { pos: None })
        }
    };

    schemapb::Type {
        value: Some(value),
    }
}
