use convert_case::{Case, Casing};
use lazy_static::lazy_static;
use regex::Regex;
use std::sync::Mutex;

struct Inflection {
    regex: Regex,
    replace: String,
}

impl Clone for Inflection {
    fn clone(&self) -> Self {
        Inflection {
            regex: Regex::new(self.regex.as_str()).unwrap(),
            replace: self.replace.clone(),
        }
    }
}

#[derive(Clone)]
struct Regular {
    find: String,
    replace: String,
}

#[derive(Clone)]
struct Irregular {
    singular: String,
    plural: String,
}

type RegularSlice = Vec<Regular>;
type IrregularSlice = Vec<Irregular>;

lazy_static! {
    static ref SINGULAR_INFLECTIONS: Mutex<RegularSlice> = Mutex::new(vec![
        // Special cases
        Regular { find: r"(vert|ind)ices$".to_string(), replace: "${1}ex".to_string() },
        Regular { find: r"(matr)ices$".to_string(), replace: "${1}ix".to_string() },
        Regular { find: r"(quiz)zes$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"(database)s$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"(drive)s$".to_string(), replace: "${1}".to_string() },
        
        // More specific patterns
        Regular { find: r"([^aeiou])ies$".to_string(), replace: "${1}y".to_string() },
        Regular { find: r"(x|ch|ss|sh|zz)es$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"([^aeiou])ys$".to_string(), replace: "${1}y".to_string() },
        Regular { find: r"(ss)$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"(n)ews$".to_string(), replace: "${1}ews".to_string() },
        Regular { find: r"([ti])a$".to_string(), replace: "${1}um".to_string() },
        Regular { find: r"((a)naly|(b)a|(d)iagno|(p)arenthe|(p)rogno|(s)ynop|(t)he)(sis|ses)$".to_string(), replace: "${1}sis".to_string() },
        Regular { find: r"(^analy)(sis|ses)$".to_string(), replace: "${1}sis".to_string() },
        Regular { find: r"([^f])ves$".to_string(), replace: "${1}fe".to_string() },
        Regular { find: r"(hive)s$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"(tive)s$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"([lr])ves$".to_string(), replace: "${1}f".to_string() },
        Regular { find: r"([^aeiouy]|qu)ies$".to_string(), replace: "${1}y".to_string() },
        Regular { find: r"(s)eries$".to_string(), replace: "${1}eries".to_string() },
        Regular { find: r"(m)ovies$".to_string(), replace: "${1}ovie".to_string() },
        Regular { find: r"^(m|l)ice$".to_string(), replace: "${1}ouse".to_string() },
        Regular { find: r"(bus)(es)?$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"(o)es$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"(shoe)s$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"(cris|test)(is|es)$".to_string(), replace: "${1}is".to_string() },
        Regular { find: r"^(a)x[ie]s$".to_string(), replace: "${1}xis".to_string() },
        Regular { find: r"(octop|vir)(us|i)$".to_string(), replace: "${1}us".to_string() },
        Regular { find: r"(alias|status)(es)?$".to_string(), replace: "${1}".to_string() },
        Regular { find: r"^(ox)en".to_string(), replace: "${1}".to_string() },
        
        // Generic patterns
        Regular { find: r"s$".to_string(), replace: "".to_string() },
    ]);

    static ref IRREGULAR_INFLECTIONS: Mutex<IrregularSlice> = Mutex::new(vec![
        Irregular { singular: "person".to_string(), plural: "people".to_string() },
        Irregular { singular: "man".to_string(), plural: "men".to_string() },
        Irregular { singular: "child".to_string(), plural: "children".to_string() },
        Irregular { singular: "sex".to_string(), plural: "sexes".to_string() },
        Irregular { singular: "move".to_string(), plural: "moves".to_string() },
        Irregular { singular: "goose".to_string(), plural: "geese".to_string() },
        Irregular { singular: "foot".to_string(), plural: "feet".to_string() },
        Irregular { singular: "tooth".to_string(), plural: "teeth".to_string() },
        Irregular { singular: "leaf".to_string(), plural: "leaves".to_string() },
        Irregular { singular: "mouse".to_string(), plural: "mice".to_string() },
        Irregular { singular: "index".to_string(), plural: "indices".to_string() },
        Irregular { singular: "axis".to_string(), plural: "axes".to_string() },
        Irregular { singular: "vertex".to_string(), plural: "vertices".to_string() },
        Irregular { singular: "matrix".to_string(), plural: "matrices".to_string() },
    ]);

    static ref UNCOUNTABLE_INFLECTIONS: Mutex<Vec<String>> = Mutex::new(vec![
        "equipment".to_string(),
        "information".to_string(),
        "rice".to_string(),
        "money".to_string(),
        "species".to_string(),
        "series".to_string(),
        "fish".to_string(),
        "sheep".to_string(),
        "jeans".to_string(),
        "police".to_string(),
        "milk".to_string(),
        "salt".to_string(),
        "time".to_string(),
        "water".to_string(),
        "paper".to_string(),
        "food".to_string(),
        "art".to_string(),
        "cash".to_string(),
        "music".to_string(),
        "help".to_string(),
        "luck".to_string(),
        "oil".to_string(),
        "progress".to_string(),
        "rain".to_string(),
        "research".to_string(),
        "shopping".to_string(),
        "software".to_string(),
        "traffic".to_string(),
        "data".to_string(),
        "metadata".to_string(),
        "status".to_string(),
        "campus".to_string(),
        "meta".to_string(),
    ]);

    static ref COMPILED_SINGULAR_MAPS: Mutex<Vec<Inflection>> = Mutex::new(Vec::new());
}

pub fn compile() {
    let mut compiled_singular_maps = COMPILED_SINGULAR_MAPS.lock().unwrap();
    compiled_singular_maps.clear();

    let irregulars = IRREGULAR_INFLECTIONS.lock().unwrap();
    for value in irregulars.iter() {
        let regex = Regex::new(&format!(r"(?i)^{}$", regex::escape(&value.plural))).unwrap();
        compiled_singular_maps.push(Inflection {
            regex,
            replace: value.singular.clone(),
        });
    }

    let uncountables = UNCOUNTABLE_INFLECTIONS.lock().unwrap();
    for uncountable in uncountables.iter() {
        let regex = Regex::new(&format!(r"(?i)^{}$", regex::escape(uncountable))).unwrap();
        compiled_singular_maps.push(Inflection {
            regex,
            replace: uncountable.clone(),
        });
    }

    let singulars = SINGULAR_INFLECTIONS.lock().unwrap();
    for value in singulars.iter() {
        let regex = Regex::new(&format!(r"(?i){}", value.find)).unwrap();
        compiled_singular_maps.push(Inflection {
            regex,
            replace: value.replace.clone(),
        });
    }
}

/// Converts a word to its singular form.
pub fn singularize(word: &str) -> String {
    lazy_static::initialize(&INITIALIZER);

    if word.is_empty() {
        return word.to_string();
    }

    let compiled_maps = COMPILED_SINGULAR_MAPS.lock().unwrap();
    for inflection in compiled_maps.iter() {
        if inflection.regex.is_match(word) {
            return inflection.regex.replace(word, &inflection.replace).to_string();
        }
    }

    word.to_string()
}

pub fn singularize_pascal(string: &str) -> String {
    singularize(string).to_case(Case::Pascal)
}

lazy_static! {
    static ref INITIALIZER: () = {
        compile();
    };
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_singularize() {
        let test_cases = vec![
            ("users", "user"),
            ("all_types", "all_type"),
            ("databases", "database"),
            ("people", "person"),
            ("metadata", "metadata"),
            ("status", "status"),
            ("indices", "index"),
            ("matrices", "matrix"),
            ("vertices", "vertex"),
            ("axes", "axis"),
            ("test_cases", "test_case"),
            ("user_addresses", "user_address"),
        ];

        for (input, expected) in test_cases {
            assert_eq!(singularize(input), expected, "Failed to singularize: {}", input);
        }
    }
}
