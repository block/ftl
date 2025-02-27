---
sidebar_position: 16
title: External Types
description: Using external types in your modules
---

# External Types

FTL supports the use of external types in your FTL modules. External types are types defined in other packages or modules that are not part of the FTL module.

The primary difference is that external types are not defined in the FTL schema, and therefore serialization and deserialization of these types is not handled 
by FTL. Instead, FTL relies on the runtime to handle serialization and deserialization of these types.

In some cases this feature can also be used to provide custom serialization and deserialization logic for types that are not directly supported by FTL, even
if they are defined in the same package as the FTL module.

When using external types:
- The external type is typically widened to `Any` in the FTL schema by default
- You can map to specific FTL types (like `String`) for better schema clarity
- For JVM languages, `java` is always used as the runtime name in the schema, regardless of whether Kotlin or Java is used
- Multiple `+typemap` annotations can be used to support cross-runtime interoperability

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs groupId="languages">
<TabItem value="go" label="Go" default>

To use an external type in your FTL module schema, declare a type alias over the external type:

```go
//ftl:typealias
type FtlType external.OtherType

//ftl:typealias
type FtlType2 = external.OtherType
```

You can also specify mappings for other runtimes:

```go
//ftl:typealias
//ftl:typemap java "com.external.other.OtherType"
type FtlType external.OtherType
```

</TabItem>
<TabItem value="kotlin" label="Kotlin">

To use an external type in your FTL module schema, implement a `TypeAliasMapper`:

```kotlin 
@TypeAlias(name = "OtherType")
class OtherTypeTypeMapper : TypeAliasMapper<OtherType, JsonNode> {
    override fun encode(`object`: OtherType): JsonNode {
        return TextNode.valueOf(`object`.value)
    }

    override fun decode(serialized: JsonNode): OtherType {
        if (serialized.isTextual) {
            return OtherType(serialized.textValue())
        }
        throw RuntimeException("Expected a textual value")
    }
}
```

Note that for JVM languages `java` is always used as the runtime name, regardless of the actual language used.

It is also possible to map to any other valid FTL type (e.g. `String`) by using this as the second type parameter:

```kotlin 
@TypeAlias(name = "OtherType")
class OtherTypeTypeMapper : TypeAliasMapper<OtherType, String> {
    override fun encode(other: OtherType): JsonNode {
        return other.value
    }

    override fun decode(serialized: String): OtherType {
        return OtherType(serialized.textValue())
    }
}
```

You can also specify mappings for other runtimes:

```kotlin
@TypeAlias(
  name = "OtherType",
  languageTypeMappings = [LanguageTypeMapping(language = "go", type = "github.com/external.OtherType")]
)
```

</TabItem>
<TabItem value="java" label="Java">

To use an external type in your FTL module schema, implement a `TypeAliasMapper`:

```java
@TypeAlias(name = "OtherType")
public class OtherTypeTypeMapper implements TypeAliasMapper<OtherType, JsonNode> {
    @Override
    public JsonNode encode(OtherType object) {
        return TextNode.valueOf(object.getValue());
    }

    @Override
    public AnySerializedType decode(OtherType serialized) {
        if (serialized.isTextual()) {
            return new OtherType(serialized.textValue());
        }
        throw new RuntimeException("Expected a textual value");
    }
}
```

It is also possible to map to any other valid FTL type (e.g. `String`) by using this as the second type parameter:

```java
@TypeAlias(name = "OtherType")
public class OtherTypeTypeMapper implements TypeAliasMapper<OtherType, String> {
    @Override
    public String encode(OtherType object) {
        return object.getValue();
    }

    @Override
    public String decode(OtherType serialized) {
        return new OtherType(serialized.textValue());
    }
}
```

You can also specify mappings for other runtimes:

```java
@TypeAlias(name = "OtherType", languageTypeMappings = {
    @LanguageTypeMapping(language = "go", type = "github.com/external.OtherType"),
})
```

</TabItem>
<TabItem value="schema" label="Schema">

In the FTL schema, external types are represented as type aliases with the `+typemap` annotation:

```
module example {
  // External type widened to Any
  typealias FtlType Any
    +typemap go "github.com/external.OtherType"
    +typemap java "foo.bar.OtherType"
  
  // External type mapped to a specific FTL type
  typealias UserID String
    +typemap go "github.com/myapp.UserID"
    +typemap java "com.myapp.UserID"
  
  // Using external types in data structures
  data User {
    id example.UserID
    preferences example.FtlType
  }
  
  // Using external types in verbs
  verb processUser(example.User) Unit
}
```

The `+typemap` annotation specifies:
1. The target runtime/language (go, java, etc.)
2. The fully qualified type name in that runtime

This allows FTL to decode the type properly in different languages, enabling seamless interoperability across different runtimes.

</TabItem>
</Tabs>
