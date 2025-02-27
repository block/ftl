---
sidebar_position: 2
title: Types
description: Declaring and using Types
---

# Types

FTL supports the following types: `Int` (64-bit), `Float` (64-bit), `String`, `Bytes` (a byte array), `Bool`, `Time`,
`Any` (a dynamic type), `Unit` (similar to "void"), arrays, maps, data structures, and constant enumerations. Each FTL
type is mapped to a corresponding language-specific type. For example in Go `Float` is represented as `float64`, `Time`
is represented by `time.Time`, and so on.

User-defined types referenced by a verb will be automatically exported as FTL types.

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Basic types

The following table shows how FTL types map to language-specific types:

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

| FTL        | Go                          |
| :--------- | :-------------------------- |
| `Int`      | `int`                       |
| `Float`    | `float64`                   |
| `String`   | `string`                    |
| `Bytes`    | `[]byte`                    |
| `Bool`     | `bool`                      |
| `Time`     | `time.Time`                 |
| `Any`      | [External](./externaltypes) |
| `Unit`     | N/A                         |
| `Map<K,V>` | `map[K]V`                   |
| `Array<T>` | `[]T`                       |

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">


| FTL        | Kotlin                      |
| :--------- | :-------------------------- |
| `Int`      | `Long`                      |
| `Float`    | `Double`                    |
| `String`   | `String`                    |
| `Bytes`    | `ByteArray`                 |
| `Bool`     | `Boolean`                   |
| `Time`     | `ZonedDateTime`             |
| `Any`      | [External](./externaltypes) |
| `Unit`     | `Unit`                      |
| `Map<K,V>` | `Map<K,V>`                  |
| `Array<T>` | `List<T>`                   |

  </TabItem>
  <TabItem value="java" label="Java">

| FTL        | Java                        | Java (optional)             |
| :--------- | :-------------------------- | :-------------------------- |
| `Int`      | `long`                      | `Long`                      |
| `Float`    | `double`                    | `Double`                    |
| `String`   | `String`                    | `@Nullable String`          |
| `Bytes`    | `[]byte`                    | `@Nullable byte[]`          |
| `Bool`     | `boolean`                   | `Boolean`                   |
| `Time`     | `ZonedDateTime️`             | `@Nullable ZonedDateTime`   |
| `Any`      | [External](./externaltypes) | [External](./externaltypes) |
| `Unit`     | `void`                      | N/A                         |
| `Map<K,V>` | `Map<K,V>`                  | `@Nullable Map<K,V>`        |
| `Array<T>` | `List<T>`                   | `@Nullable List<T>`         |

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, these are the primitive types that can be used directly:

| FTL Type   | Description                                |
| :--------- | :----------------------------------------- |
| `Int`      | 64-bit signed integer                      |
| `Float`    | 64-bit floating point number               |
| `String`   | UTF-8 encoded string                       |
| `Bytes`    | Byte array                                 |
| `Bool`     | Boolean value (true/false)                 |
| `Time`     | RFC3339 formatted timestamp                |
| `Any`      | Dynamic type (schema-less)                 |
| `Unit`     | Empty type (similar to void)               |
| `Map<K,V>` | Key-value mapping                          |
| `[T]`      | Array of type T                            |

Example usage in schema:
```
module example {
  data Person {
    name String
    age Int
    height Float
    isActive Bool
    birthdate Time
    metadata Map<String, Any>
    tags [String]
  }
}
```

  </TabItem>
</Tabs>

## Data structures

FTL supports user-defined data structures, declared using the idiomatic syntax of the target language.

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```go
type Person struct {
  Name string
  Age  int
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
data class Person(
  val name: String,
  val age: Int
)
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
public class Person {
  private final String name;
  private final int age;

  public Person(String name, int age) {
    this.name = name;
    this.age = age;
  }
}
```

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, data structures are defined using the `data` keyword:

```
module example {
  data Person {
    name String
    age Int
  }
}
```

Fields can have optional values by adding a `?` suffix:

```
module example {
  data Person {
    name String
    age Int
    address String?
  }
}
```

  </TabItem>
</Tabs>

## Generics

FTL has first-class support for generics, declared using the idiomatic syntax of the target language.

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```go
type Pair[T, U] struct {
  First  T
  Second U
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
data class Pair<T, U>(
  val first: T,
  val second: U
)
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
public class Pair<T, U> {
  private final T first;
  private final U second;

  public Pair(T first, U second) {
    this.first = first;
    this.second = second;
  }
}
```

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, generic types are defined with type parameters:

```
module example {
  data Pair<T, U> {
    first T
    second U
  }
}
```

Generic types can be used in other type definitions:

```
module example {
  data Pair<T, U> {
    first T
    second U
  }
  
  data StringIntPair {
    pair example.Pair<String, Int>
  }
}
```

  </TabItem>
</Tabs>

## Type enums (sum types)

[Sum types](https://en.wikipedia.org/wiki/Tagged_union) are supported by FTL's type system.

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

Sum types aren't directly supported by Go, however they can be approximated with the use of [sealed interfaces](https://blog.chewxy.com/2018/03/18/golang-interfaces/):

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}

type Dog struct {}
func (Dog) animal() {}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

Sum types aren't directly supported by Kotlin, however they can be approximated with the use of [sealed interfaces](https://kotlinlang.org/docs/sealed-classes.html):

```kotlin
@Enum
sealed interface Animal

@EnumHolder
class Cat() : Animal

@EnumHolder
class Dog() : Animal
```

  </TabItem>
  <TabItem value="java" label="Java">

> TODO

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, sum types (type enums) are represented as a union of types:

```
module example {
  data Cat {}
  
  data Dog {}
  
  enum Animal {
    Cat example.Cat
    Dog example.Dog
  }
}
```

When used in other types or verbs, the sum type can be referenced directly:

```
module example {
  verb processAnimal(example.Animal) Unit
}
```

  </TabItem>
</Tabs>

## Value enums

A value enum is an enumerated set of string or integer values.

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```go
//ftl:enum
type Colour string

const (
  Red   Colour = "red"
  Green Colour = "green"
  Blue  Colour = "blue"
)

//ftl:enum
type Status int

const (
  Active   Status = 1
  Inactive Status = 0
  Pending  Status = 2
)
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
@Enum
public enum class Colour(
  public final val `value`: String,
) {
  Red("red"),
  Green("green"),
  Blue("blue"),
  ;
}

@Enum
public enum class Status(
  public final val `value`: Int,
) {
  Active(1),
  Inactive(0),
  Pending(2),
  ;
}
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
@Enum
public enum Colour {
  Red("red"),
  Green("green"),
  Blue("blue");

  private final String value;

  Colour(String value) {
    this.value = value;
  }
}

@Enum
public enum Status {
  Active(1),
  Inactive(0),
  Pending(2);

  private final int value;

  Status(int value) {
    this.value = value;
  }
}
```

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, value enums are represented as an enum with string or integer values:

```
module example {
  enum Colour: String {
    Red = "red"
    Green = "green"
    Blue = "blue"
  }
  
  enum Status: Int {
    Active = 1
    Inactive = 0
    Pending = 2
  }
}
```

  </TabItem>
</Tabs>

## Type aliases

A type alias is an alternate name for an existing type. It can be declared like so:

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```go
//ftl:typealias
type UserID string
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
typealias UserID = String
```

  </TabItem>
  <TabItem value="java" label="Java">

```java
// Java does not support type aliases directly
// Use a wrapper class instead
public class UserID {
    private final String value;

    public UserID(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }
}
```

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, type aliases are defined using the `typealias` keyword:

```
module example {
  typealias UserID String
}
```

Type aliases can be used in data structures:

```
module example {
  typealias UserID String
  
  typealias UserMap Map<String, example.User>
  
  data User {
    id example.UserID
    name String
  }
}
```

  </TabItem>
</Tabs>

Type aliases are useful for making code more readable and type-safe by giving meaningful names to types that represent specific concepts in your domain.
