use std::collections::{HashMap, HashSet, VecDeque};
use std::path::{Path, PathBuf};

use proc_macro2::{Ident, Span};
use tracing::{debug, warn};

use ftl_protos::schema;

#[derive(Debug)]
pub struct Module {
    pub name: ModuleIdent,
    pub ast: syn::File,
}

pub struct Parser {
    modules: Vec<Module>,
}

impl Parser {
    pub fn new() -> Self {
        Self {
            modules: Vec::new(),
        }
    }

    pub fn add_module(&mut self, module_ident: &ModuleIdent, code: &str) {
        let ast = syn::parse_file(code).unwrap();
        self.modules.push(Module {
            name: module_ident.clone(),
            ast,
        });
    }

    pub fn add_glob(&mut self, path: &Path) {
        let sources = find_sources(path);
        for source in sources {
            let contents = std::fs::read_to_string(&source).unwrap();
            let file_name = source.file_stem().unwrap().to_str().unwrap();
            let mod_ident = ModuleIdent::new(file_name);

            self.add_module(&mod_ident, &contents);
        }
    }

    pub fn parse(self) -> Parsed {
        // Get all verbs.
        let verbs: Vec<_> = self
            .modules
            .iter()
            .flat_map(|module| extract_ast_verbs(&module.name, &module.ast))
            .collect();

        println!("Found {} verbs", verbs.len());

        // For each verb, queue request and response types.
        let mut type_queue = VecDeque::new();
        for verb in &verbs {
            let request_ident = ident_from_type(&verb.request_type);
            let response_ident = ident_from_type(&verb.response_type);
            type_queue.push_back(request_ident.clone());
            type_queue.push_back(response_ident.clone());
        }

        // Iterate queue until empty.
        let mut types = HashMap::new(); // Ident to AST node
        while let Some(ident) = type_queue.pop_front() {
            let (item, add_to_queue) = self.find_type_ast(&ident);
            if let Some(item) = item {
                types.insert(ident.clone(), item);
            }
            for ident in add_to_queue {
                type_queue.push_back(ident);
            }
            println!("Types: {:?}", types);
            println!("Queue: {:?}", type_queue);
        }

        Parsed { verbs, types }
    }

    fn find_type_ast(&self, ident: &Ident) -> (Option<syn::Item>, Vec<syn::Ident>) {
        let known_types = [
            "String", "bool", "u8", "u16", "u32", "u64", "i8", "i16", "i32", "i64",
        ];

        if known_types.contains(&ident.to_string().as_str()) {
            return (None, vec![]);
        }

        for module in &self.modules {
            for item in &module.ast.items {
                let syn::Item::Struct(item_struct) = item else {
                    continue;
                };

                if item_struct.ident != *ident {
                    continue;
                }

                // Get the ident of the Type for each field.
                let fields = item_struct
                    .fields
                    .iter()
                    .map(|field| ident_from_type(&field.ty))
                    .collect();

                return (Some(item.clone()), fields);
            }
        }

        panic!("Type not found: {:?}", ident);
    }
}

fn ident_from_path(path: &syn::Path) -> Ident {
    path.get_ident().unwrap().clone()
}

fn ident_from_type(ty: &syn::Type) -> Ident {
    match ty {
        syn::Type::Path(path) => ident_from_path(&path.path),
        _ => panic!("Expected type path, got {:?}", ty),
    }
}

pub struct Parsed {
    pub verbs: Vec<VerbToken>,
    types: HashMap<Ident, syn::Item>,
}

impl Parsed {
    pub fn modules_count(&self) -> usize {
        self.verbs
            .iter()
            .map(|verb| &verb.module)
            .collect::<HashSet<_>>()
            .len()
    }

    pub fn verb_count(&self) -> usize {
        self.verbs.len()
    }
}

/// Find all .rs sources in src.
fn find_sources(path: &Path) -> Vec<PathBuf> {
    glob::glob(path.to_str().unwrap())
        .unwrap()
        .map(|entry| entry.unwrap())
        .map(|entry| entry.as_path().to_path_buf())
        .collect()
}

#[derive(Debug, Clone, Hash, Eq, PartialEq)]
pub struct ModuleIdent(pub Ident);

impl ModuleIdent {
    pub fn new(name: &str) -> Self {
        Self(Ident::new(name, Span::call_site()))
    }
}

#[derive(Debug, Clone)]
pub struct VerbToken {
    pub module: ModuleIdent,
    pub func: syn::ItemFn,
    pub request_type: syn::Type,
    pub response_type: syn::Type,
}

impl VerbToken {
    pub fn try_parse_any_item(module: &ModuleIdent, item: &syn::Item) -> Option<Self> {
        let func = match item {
            syn::Item::Fn(func) => func,
            // Quietly ignore non-functions.
            _ => return None,
        };

        let ftl_verb_path = syn::parse_str("ftl::verb").unwrap();
        if !has_meta_path(&func.attrs, &ftl_verb_path) {
            // No #[ftl::verb] annotation, ignore.
            return None;
        }

        ensure_arg_is_path(&func, 0);
        ensure_arg_type_ident(&func, 0, "Context");
        ensure_arg_is_path(&func, 1);
        let request_type = get_arg_type(&func, 1);
        let response_type = get_return_type_or_unit(&func);

        Some(VerbToken {
            module: module.clone(),
            func: func.clone(),
            request_type,
            response_type,
        })
    }

    pub fn to_proto(&self) -> schema::Verb {
        let mut verb = schema::Verb::default();
        verb.name = self.func.sig.ident.to_string();

        warn!("TODO: for now just give unit types");
        let request = Some(schema::Type {
            value: Some(schema::r#type::Value::Unit(schema::Unit { pos: None })),
        });
        let response = Some(schema::Type {
            value: Some(schema::r#type::Value::Unit(schema::Unit { pos: None })),
        });

        schema::Verb {
            runtime: None,
            pos: None,
            comments: vec![],
            export: false,
            name: self.func.sig.ident.to_string(),
            request,
            response,
            metadata: vec![],
        }
    }
}

/// Extract functions that are annotated with #[ftl::verb] and extract the AST node.
pub fn extract_ast_verbs(module: &ModuleIdent, ast: &syn::File) -> Vec<VerbToken> {
    let mut verbs = vec![];
    for item in &ast.items {
        let Some(verb_token) = VerbToken::try_parse_any_item(&module, &item) else {
            continue;
        };

        verbs.push(verb_token);
    }
    verbs
}

fn ensure_arg_is_path(func: &syn::ItemFn, index: usize) {
    if func.sig.inputs.len() <= index {
        panic!("Function must have at least {} arguments", index + 1);
    }
    let arg = &func.sig.inputs[index];

    let syn::FnArg::Typed(pat) = arg else {
        panic!("Argument {} must not be a self argument", index);
    };

    let syn::Type::Path(path) = &*pat.ty else {
        panic!(
            "Argument {} must be of type Path instead of {:?}",
            index, pat.ty
        );
    };
}

fn ensure_arg_type_ident(func: &syn::ItemFn, index: usize, ident: &str) {
    if func.sig.inputs.len() <= index {
        panic!("Function must have at least {} arguments", index + 1);
    }
    let arg = &func.sig.inputs[index];

    let syn::FnArg::Typed(pat) = arg else {
        panic!("Argument {} must not be a self argument", index);
    };

    let syn::Type::Path(path) = &*pat.ty else {
        panic!(
            "Argument {} must be of type Path instead of {:?}",
            index, pat.ty
        );
    };

    let path = &path.path else {
        panic!(
            "Argument {} must be of type Path instead of {:?}",
            index, path.path
        );
    };
}

fn get_arg_type(func: &syn::ItemFn, index: usize) -> syn::Type {
    if func.sig.inputs.len() <= index {
        panic!("Function must have at least {} arguments", index + 1);
    }
    let arg = &func.sig.inputs[index];

    let syn::FnArg::Typed(pat) = arg else {
        panic!("Argument {} must not be a self argument", index);
    };

    *pat.ty.clone()
}

fn get_return_type_or_unit(func: &syn::ItemFn) -> syn::Type {
    let syn::ReturnType::Type(_, ty) = &func.sig.output else {
        return syn::Type::Tuple(syn::TypeTuple {
            paren_token: syn::token::Paren::default(),
            elems: Default::default(),
        });
    };

    *ty.clone()
}

// Look for #[path_str] e.g. #[ftl::verb], and extract the function signature.
fn has_meta_path(attrs: &[syn::Attribute], expected_path: &syn::Path) -> bool {
    attrs.iter().any(|attr| attr.meta.path() == expected_path)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_types() {
        let mut parser = Parser::new();
        parser.add_module(
            &ModuleIdent::new("test"),
            r#"
            pub struct User {
                pub name: String,
                pub age: u32,
            }

            pub struct Request {
                pub name: String,
                pub user: User,
            }

            pub struct Response {
                pub message: String,
                pub user: User,
            }

            #[ftl::verb]
            pub fn test(ctx: Context, request: Request) -> Response {
                println!("Hello, world!");
            }
        "#,
        );
        let parsed = parser.parse();

        assert_eq!(parsed.modules_count(), 1);
        assert_eq!(parsed.verb_count(), 1);
    }
}
