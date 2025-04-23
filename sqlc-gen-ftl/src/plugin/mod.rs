#![allow(dead_code)]

mod inflection;

use crate::protos::pluginpb;
use crate::protos::schemapb;
use crate::protos::schemapb::r#type::Value as TypeValue;
use convert_case::{Case, Casing};
use prost::Message;
use std::io;
use std::collections::HashMap;
use self::inflection::singularize_pascal;
use std::collections::HashSet;

pub struct Plugin;

impl Plugin {
    pub fn generate_from_input(input: &[u8]) -> Result<Vec<u8>, io::Error> {
        let req = pluginpb::GenerateRequest::decode(input)
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))?;
        let resp = Self::handle_generate(req)?;
        Ok(resp.encode_to_vec())
    }

    fn handle_generate(
        req: pluginpb::GenerateRequest
    ) -> Result<pluginpb::GenerateResponse, io::Error> {
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
    let (module_name, _) = get_options(request)?;
    
    let mut table_schemas = HashMap::new();
    for query in &request.queries {
        if let Some(table_name) = get_table(query) {
            let pascal_singular = singularize_pascal(&table_name);
            // Add if we haven't seen this table before and the query represents a full table schema
            // e.g. `SELECT * FROM users` will produce a table schema for `users` and ultimately a `User` schema type
            if !table_schemas.contains_key(&pascal_singular) {
                if is_full_table_schema(query, request) {
                    let table_decl = create_table_schema(query, &pascal_singular, request);
                    table_schemas.insert(pascal_singular.clone(), table_decl);
                }
            }
        }
    }
    
    let mut table_decls = Vec::new();
    for (_, decl) in table_schemas {
        table_decls.push(decl);
    }
    decls.extend(table_decls);
    
    for query in &request.queries {
        // For queries with parameters, create request types as needed.
        // If there's only one parameter, expose its type directly.
        if !query.params.is_empty() {
            let request_decl = to_request_type(query, request, &decls);
            if let Some(decl) = request_decl {
                decls.push(decl);
            }
        }

        // For queries with result columns, create response types as needed.
        // If there's only one column, expose its type directly.
        if !query.columns.is_empty() {
            let response_decl = to_response_type(query, request, &decls);
            if let Some(decl) = response_decl {
                decls.push(decl);
            }
        }

        decls.push(to_verb(query, request, &decls));
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

fn is_full_table_schema(query: &pluginpb::Query, request: &pluginpb::GenerateRequest) -> bool {
    if let Some(table_name) = get_table(query) {
        let table_columns = get_table_columns(&table_name, request);
        
        if table_columns.is_empty() {
            return false;
        }
        
        let all_from_same_table = query.columns.iter().all(|col| {
            match &col.table {
                Some(table) => table.name == table_name,
                None => false,
            }
        });
        
        if !all_from_same_table {
            return false;
        }
        
        let query_column_names: Vec<String> = query.columns
            .iter()
            .map(|col| col.name.clone())
            .collect();
            
        if query_column_names.len() != table_columns.len() {
            return false;
        }
        
        for table_col in &table_columns {
            if !query_column_names.contains(&table_col.name) {
                return false;
            }
        }
        
        return true;
    }
    false
}

fn get_table_columns(table_name: &str, request: &pluginpb::GenerateRequest) -> Vec<pluginpb::Column> {
    let columns = Vec::new();
    if let Some(catalog) = &request.catalog {
        for schema in &catalog.schemas {
            for table in &schema.tables {
                if let Some(rel) = &table.rel {
                    if rel.name == table_name {
                        return table.columns.clone();
                    }
                }
            }
        }
    }
    columns
}

fn get_table(query: &pluginpb::Query) -> Option<String> {
    if let Some(insert_table) = &query.insert_into_table {
        if !insert_table.name.is_empty() {
            return Some(insert_table.name.clone());
        }
    }
    if !query.columns.is_empty() {
        for col in &query.columns {
            if let Some(table) = &col.table {
                if !table.name.is_empty() {
                    return Some(table.name.clone());
                }
            }
        }
    }
    None
}

fn create_table_schema(
    query: &pluginpb::Query, 
    name: &str, 
    request: &pluginpb::GenerateRequest
) -> schemapb::Decl {
    schemapb::Decl {
        value: Some(
            schemapb::decl::Value::Data(schemapb::Data {
                name: name.to_string(),
                export: false,
                type_parameters: Vec::new(),
                fields: query.columns
                    .iter()
                    .map(|col| {
                        to_schema_field(col.name.clone(), Some(col), col.r#type.as_ref(), request)
                    })
                    .collect(),
                pos: None,
                comments: Vec::new(),
                metadata: vec![generated_metadata()],
            })
        ),
    }
}

fn params_match_table(
    query: &pluginpb::Query, 
    table_name: &str, 
    request: &pluginpb::GenerateRequest
) -> bool {
    let table_columns = get_table_columns(table_name, request);
    if table_columns.is_empty() {
        return false;
    }
    
    if query.params.len() != table_columns.len() {
        return false;
    }
    
    let mut param_matches = 0;
    for param in &query.params {
        if let Some(col) = &param.column {
            if let Some(col_table) = &col.table {
                if col_table.name == table_name {
                    if table_columns.iter().any(|tc| tc.name == col.name) {
                        param_matches += 1;
                    }
                }
            }
        }
    }
    
    param_matches == query.params.len() && param_matches == table_columns.len()
}

fn to_request_type(
    query: &pluginpb::Query, 
    request: &pluginpb::GenerateRequest,
    decls: &[schemapb::Decl]
) -> Option<schemapb::Decl> {
    if query.params.len() == 1 {
        return None;
    }
    
    if let Some(table_name) = get_table(query) {
        let pascal_singular = singularize_pascal(&table_name);
        let table_schema_exists = decls.iter().any(|decl| {
            if let Some(schemapb::decl::Value::Data(data)) = &decl.value {
                data.name == pascal_singular
            } else {
                false
            }
        });
        
        if table_schema_exists && params_match_table(query, &table_name, request) {
            return None; // Use the existing table schema.
        }
    }
    
    let upper_camel_name = to_upper_camel(&query.name);
    let mut used_field_names = HashSet::new();
    Some(schemapb::Decl {
        value: Some(
            schemapb::decl::Value::Data(schemapb::Data {
                name: format!("{}Query", upper_camel_name),
                export: false,
                type_parameters: Vec::new(),
                fields: query.params
                    .iter()
                    .map(|param| {
                        let base_name = if let Some(col) = &param.column {
                            col.name.clone()
                        } else {
                            format!("arg{}", param.number)
                        };

                        // Find a unique name by adding numbers if needed
                        let mut name = base_name.clone();
                        let mut counter = param.number;
                        while !used_field_names.insert(name.clone()) {
                            name = format!("{}{}", base_name, counter);
                            counter += 1;
                        }

                        let sql_type = param.column.as_ref().and_then(|col| col.r#type.as_ref());
                        to_schema_field(name, param.column.as_ref(), sql_type, request)
                    })
                    .collect(),
                pos: None,
                comments: Vec::new(),
                metadata: vec![generated_metadata()],
            })
        ),
    })
}

fn to_response_type(
    query: &pluginpb::Query, 
    request: &pluginpb::GenerateRequest,
    decls: &[schemapb::Decl]
) -> Option<schemapb::Decl> {
    if query.columns.len() == 1 {
        return None;
    }
    
    if let Some(table_name) = get_table(query) {
        let pascal_singular = singularize_pascal(&table_name);
        let table_schema_exists = decls.iter().any(|decl| {
            if let Some(schemapb::decl::Value::Data(data)) = &decl.value {
                data.name == pascal_singular
            } else {
                false
            }
        });
        
        if table_schema_exists && is_full_table_schema(query, request) {
            return None; // Use the existing table schema.
        }
    }
    
    let pascal_name = to_upper_camel(&query.name);
    let struct_name = format!("{}Row", pascal_name);
    
    Some(schemapb::Decl {
        value: Some(
            schemapb::decl::Value::Data(schemapb::Data {
                name: struct_name,
                export: false,
                type_parameters: Vec::new(),
                fields: query.columns
                    .iter()
                    .map(|col| {
                        to_schema_field(col.name.clone(), Some(col), col.r#type.as_ref(), request)
                    })
                    .collect(),
                pos: None,
                comments: Vec::new(),
                metadata: vec![generated_metadata()],
            })
        ),
    })
}

fn to_verb(
    query: &pluginpb::Query, 
    request: &pluginpb::GenerateRequest,
    decls: &[schemapb::Decl]
) -> schemapb::Decl {
    let (module_name, database_name) = get_options(request).unwrap();
    let upper_camel_name = to_upper_camel(&query.name);
    
    let request_type = if query.params.is_empty() {
        if query.insert_into_table.is_some() {
            if let Some(table_name) = get_table(query) {
                let pascal_singular = singularize_pascal(&table_name);
                Some(to_schema_ref(&module_name, &pascal_singular))
            } else {
                Some(to_schema_unit())
            }
        } else {
            Some(to_schema_unit())
        }
    } else if query.params.len() == 1 {
        Some(get_parameter_type(&query.params[0], request))
    } else if let Some(table_name) = get_table(query) {
        let pascal_singular = singularize_pascal(&table_name);
        let table_schema_exists = decls.iter().any(|decl| {
            if let Some(schemapb::decl::Value::Data(data)) = &decl.value {
                data.name == pascal_singular
            } else {
                false
            }
        });
        
        if table_schema_exists && params_match_table(query, &table_name, request) {
            Some(to_schema_ref(&module_name, &pascal_singular))
        } else {
            Some(to_schema_ref(&module_name, &format!("{}Query", upper_camel_name)))
        }
    } else {
        Some(to_schema_ref(&module_name, &format!("{}Query", upper_camel_name)))
    };


    let response_type = match query.cmd.as_str() {
        ":exec" => Some(to_schema_unit()),
        ":one" => {
            if query.columns.len() == 1 {
                Some(get_column_type(&query.columns[0], request))
            } else if let Some(table_name) = get_table(query) {
                let pascal_singular = singularize_pascal(&table_name);
                let table_schema_exists = decls.iter().any(|decl| {
                    if let Some(schemapb::decl::Value::Data(data)) = &decl.value {
                        data.name == pascal_singular
                    } else {
                        false
                    }
                });
                
                if table_schema_exists && is_full_table_schema(query, request) {
                    Some(to_schema_ref(&module_name, &pascal_singular))
                } else {
                    Some(to_schema_ref(&module_name, &format!("{}Row", upper_camel_name)))
                }
            } else {
                Some(to_schema_ref(&module_name, &format!("{}Row", upper_camel_name)))
            }
        }
        ":many" => {
            let element_type = if query.columns.len() == 1 {
                get_column_type(&query.columns[0], request)
            } else if let Some(table_name) = get_table(query) {
                let pascal_singular = singularize_pascal(&table_name);
                let table_schema_exists = decls.iter().any(|decl| {
                    if let Some(schemapb::decl::Value::Data(data)) = &decl.value {
                        data.name == pascal_singular
                    } else {
                        false
                    }
                });
                
                if table_schema_exists && is_full_table_schema(query, request) {
                    to_schema_ref(&module_name, &pascal_singular)
                } else {
                    to_schema_ref(&module_name, &format!("{}Row", upper_camel_name))
                }
            } else {
                to_schema_ref(&module_name, &format!("{}Row", upper_camel_name))
            };
            
            Some(schemapb::Type {
                value: Some(TypeValue::Array(Box::new(schemapb::Array {
                    pos: None,
                    element: Some(Box::new(element_type)),
                })))
            })
        }
        _ => Some(to_schema_unit()),
    };

    let sql_query_metadata = schemapb::Metadata {
        value: Some(
            schemapb::metadata::Value::SqlQuery(schemapb::MetadataSqlQuery {
                pos: None,
                query: query.text
                .replace('\n', " ")
                .split_whitespace()
                .collect::<Vec<&str>>()
                .join(" "),
                command: query.cmd.trim_start_matches(':').to_string(),
            })
        ),
    };

    let db_call_metadata = schemapb::Metadata {
        value: Some(
            schemapb::metadata::Value::Databases(schemapb::MetadataDatabases {
                pos: None,
                uses: vec![schemapb::Ref {
                    pos: None,
                    module: module_name.to_string(),
                    name: database_name.to_string(),
                    type_parameters: vec![],
                }],
            })
        ),
    };

    schemapb::Decl {
        value: Some(
            schemapb::decl::Value::Verb(schemapb::Verb {
                name: query.name.to_case(Case::Camel),
                export: false,
                runtime: None,
                request: request_type,
                response: response_type,
                pos: None,
                comments: Vec::new(),
                metadata: vec![sql_query_metadata, db_call_metadata, generated_metadata()],
            })
        ),
    }
}

fn to_schema_field(
    name: String,
    col: Option<&pluginpb::Column>,
    sql_type: Option<&pluginpb::Identifier>,
    req: &pluginpb::GenerateRequest
) -> schemapb::Field {
    let mut metadata = Vec::new();

    if let Some(col) = col {
        if let Some(table) = &col.table {
            let db_column = schemapb::Metadata {
                value: Some(
                    schemapb::metadata::Value::SqlColumn(schemapb::MetadataSqlColumn {
                        pos: None,
                        table: table.name.clone(),
                        name: col.name.clone(),
                    })
                ),
            };
            metadata.push(db_column);
        }
    }

    schemapb::Field {
        name: name.to_case(Case::Camel),
        r#type: Some(match (col, sql_type) {
            (Some(col), _) => to_schema_type(req, col),
            (None, _) =>
                schemapb::Type {
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
        value: Some(
            schemapb::r#type::Value::Ref(schemapb::Ref {
                module: module_name.to_string(),
                name: name.to_string(),
                pos: None,
                type_parameters: vec![],
            })
        ),
    }
}

fn to_schema_unit() -> schemapb::Type {
    schemapb::Type {
        value: Some(
            schemapb::r#type::Value::Unit(schemapb::Unit {
                pos: None,
            })
        ),
    }
}

fn get_options(req: &pluginpb::GenerateRequest) -> Result<(String, String), io::Error> {
    let codegen = req.settings
        .as_ref()
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidData, "Missing settings"))?
        .codegen.as_ref()
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidData, "Missing codegen settings"))?;

    let options_str = String::from_utf8(codegen.options.clone()).map_err(|e|
        io::Error::new(io::ErrorKind::InvalidData, format!("Invalid UTF-8 in options: {}", e))
    )?;

    let options: serde_json::Value = serde_json::from_str(&options_str)
        .map_err(|e|
            io::Error::new(
                io::ErrorKind::InvalidData,
                format!("Failed to parse JSON options: {}", e)
            )
        )?;

    let module_name = options
        .get("module")
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
        .ok_or_else(||
            io::Error::new(io::ErrorKind::InvalidData, "Missing module name in options")
        )?;

    let database_name = options
        .get("database")
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
        .ok_or_else(||
            io::Error::new(io::ErrorKind::InvalidData, "Missing database name in options")
        )?;

    Ok((module_name, database_name))
}

fn to_upper_camel(s: &str) -> String {
    let snake = s.to_case(Case::Snake);
    snake
        .split('_')
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
        _ => mysql_to_schema_type(col),
    }
}

fn mysql_to_schema_type(col: &pluginpb::Column) -> schemapb::Type {
    let column_type = col.r#type
        .as_ref()
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
                        value: Some(TypeValue::String(schemapb::String { pos: None })),
                    })),
                }))
            }
        }
        "tinyint" => {
            if length == 1 {
                if not_null {
                    TypeValue::Bool(schemapb::Bool { pos: None })
                } else {
                    TypeValue::Optional(Box::new(schemapb::Optional {
                        pos: None,
                        r#type: Some(Box::new(schemapb::Type {
                            value: Some(TypeValue::Bool(schemapb::Bool { pos: None })),
                        })),
                    }))
                }
            } else {
                if not_null {
                    TypeValue::Int(schemapb::Int { pos: None })
                } else {
                    TypeValue::Optional(Box::new(schemapb::Optional {
                        pos: None,
                        r#type: Some(Box::new(schemapb::Type {
                            value: Some(TypeValue::Int(schemapb::Int { pos: None })),
                        })),
                    }))
                }
            }
        }
        "year" | "smallint" => {
            if not_null {
                TypeValue::Int(schemapb::Int { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Int(schemapb::Int { pos: None })),
                    })),
                }))
            }
        }
        "int" | "integer" | "mediumint" | "bigint" => {
            if not_null {
                TypeValue::Int(schemapb::Int { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Int(schemapb::Int { pos: None })),
                    })),
                }))
            }
        }
        "blob" | "binary" | "varbinary" | "tinyblob" | "mediumblob" | "longblob" => {
            if not_null {
                TypeValue::Bytes(schemapb::Bytes { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Bytes(schemapb::Bytes { pos: None })),
                    })),
                }))
            }
        }
        "double" | "double precision" | "real" | "float" => {
            if not_null {
                TypeValue::Float(schemapb::Float { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Float(schemapb::Float { pos: None })),
                    })),
                }))
            }
        }
        "decimal" | "dec" | "fixed" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None })),
                    })),
                }))
            }
        }
        "enum" => TypeValue::String(schemapb::String { pos: None }),
        "date" | "timestamp" | "datetime" | "time" => {
            if not_null {
                TypeValue::Time(schemapb::Time { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Time(schemapb::Time { pos: None })),
                    })),
                }))
            }
        }
        "boolean" | "bool" => {
            if not_null {
                TypeValue::Bool(schemapb::Bool { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Bool(schemapb::Bool { pos: None })),
                    })),
                }))
            }
        }
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
    let column_type = col.r#type
        .as_ref()
        .map(|t| t.name.to_lowercase())
        .unwrap_or_default();
    let not_null = col.not_null || col.is_array;

    let value = match column_type.as_str() {
        | "smallint"
        | "int2"
        | "pg_catalog.int2"
        | "integer"
        | "int"
        | "int4"
        | "pg_catalog.int4"
        | "bigint"
        | "int8"
        | "pg_catalog.int8"
        | "smallserial"
        | "serial2"
        | "pg_catalog.serial2"
        | "serial"
        | "serial4"
        | "pg_catalog.serial4"
        | "bigserial"
        | "serial8"
        | "pg_catalog.serial8" => {
            if not_null {
                TypeValue::Int(schemapb::Int { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Int(schemapb::Int { pos: None })),
                    })),
                }))
            }
        }

        | "float"
        | "double precision"
        | "float8"
        | "pg_catalog.float8"
        | "real"
        | "float4"
        | "pg_catalog.float4" => {
            if not_null {
                TypeValue::Float(schemapb::Float { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Float(schemapb::Float { pos: None })),
                    })),
                }))
            }
        }

        "numeric" | "decimal" | "pg_catalog.numeric" | "money" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None })),
                    })),
                }))
            }
        }

        "boolean" | "bool" | "pg_catalog.bool" => {
            if not_null {
                TypeValue::Bool(schemapb::Bool { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Bool(schemapb::Bool { pos: None })),
                    })),
                }))
            }
        }

        "json" | "jsonb" => TypeValue::Any(schemapb::Any { pos: None }),

        "bytea" | "blob" | "pg_catalog.bytea" => TypeValue::Bytes(schemapb::Bytes { pos: None }),

        | "date"
        | "pg_catalog.time"
        | "pg_catalog.timetz"
        | "pg_catalog.timestamp"
        | "pg_catalog.timestamptz"
        | "timestamptz"
        | "time"
        | "timestamp" => {
            if not_null {
                TypeValue::Time(schemapb::Time { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::Time(schemapb::Time { pos: None })),
                    })),
                }))
            }
        }

        | "text"
        | "pg_catalog.varchar"
        | "pg_catalog.bpchar"
        | "string"
        | "citext"
        | "name"
        | "uuid"
        | "ltree"
        | "lquery"
        | "ltxtquery"
        | "varchar"
        | "char"
        | "character"
        | "pg_catalog.char"
        | "pg_catalog.character"
        | "bpchar" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None })),
                    })),
                }))
            }
        }

        "inet" | "cidr" | "macaddr" | "macaddr8" => {
            if not_null {
                TypeValue::String(schemapb::String { pos: None })
            } else {
                TypeValue::Optional(Box::new(schemapb::Optional {
                    pos: None,
                    r#type: Some(Box::new(schemapb::Type {
                        value: Some(TypeValue::String(schemapb::String { pos: None })),
                    })),
                }))
            }
        }

        "box" | "circle" | "line" | "lseg" | "path" | "point" | "polygon" => {
            TypeValue::String(schemapb::String { pos: None })
        }

        "bit" | "varbit" | "pg_catalog.bit" | "pg_catalog.varbit" => {
            TypeValue::String(schemapb::String { pos: None })
        }

        | "daterange"
        | "tsrange"
        | "tstzrange"
        | "numrange"
        | "int4range"
        | "int8range"
        | "datemultirange"
        | "tsmultirange"
        | "tstzmultirange"
        | "nummultirange"
        | "int4multirange"
        | "int8multirange" => {
            TypeValue::Any(schemapb::Any { pos: None })
        }

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

fn generated_metadata() -> schemapb::Metadata {
    schemapb::Metadata {
        value: Some(
            schemapb::metadata::Value::Generated(schemapb::MetadataGenerated { pos: None })
        ),
    }
}

fn get_column_type(
    col: &pluginpb::Column,
    req: &pluginpb::GenerateRequest
) -> schemapb::Type {
    to_schema_type(req, col)
}

fn get_parameter_type(
    param: &pluginpb::Parameter,
    req: &pluginpb::GenerateRequest
) -> schemapb::Type {
    if let Some(col) = &param.column {
        to_schema_type(req, col)
    } else {
        schemapb::Type {
            value: Some(schemapb::r#type::Value::Any(schemapb::Any { pos: None }))
        }
    }
}
