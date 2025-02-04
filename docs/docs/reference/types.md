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

| FTL        | Go                          | Kotlin                      | Java                        | Java (optional)             |
| :--------- | :-------------------------- | :-------------------------- | :-------------------------- | :-------------------------- |
| `Int`      | `int`                       | `Long`                      | `long`                      | `Long`                      |
| `Float`    | `float64`                   | `Double`                    | `double`                    | `Double`                    |
| `String`   | `string`                    | `String`                    | `String`                    | `@Nullable String`          |
| `Bytes`    | `[]byte`                    | `ByteArray`                 | `[]byte`                    | `@Nullable byte[]`          |
| `Bool`     | `bool`                      | `Boolean`                   | `boolean`                   | `Boolean`                   |
| `Time`     | `time.Time`                 | `ZonedDateTime`             | `ZonedDateTime️`             | `@Nullable ZonedDateTime`   |
| `Any`      | [External](./externaltypes) | [External](./externaltypes) | [External](./externaltypes) | [External](./externaltypes) |
| `Unit`     | N/A                         | N/A                         | `void`                      | N/A                         |
| `Map<K,V>` | `map[K]V`                   | `Map<K,V>`                  | `Map<K,V>`                  | `@Nullable Map<K,V>`        |
| `Array<T>` | `[]T`                       | `List<T>`                   | `List<T>`                   | `@Nullable List<T>`         |

## Data structures

FTL supports user-defined data structures, declared using the idiomatic syntax of the target language.

<Tabs>
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
</Tabs>

## Generics

FTL has first-class support for generics, declared using the idiomatic syntax of the target language.

<Tabs>
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
</Tabs>

## Type enums (sum types)

[Sum types](https://en.wikipedia.org/wiki/Tagged_union) are supported by FTL's type system.

<Tabs>
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
</Tabs>

## Value enums

A value enum is an enumerated set of string or integer values.

<Tabs>
  <TabItem value="go" label="Go" default>

```go
//ftl:enum
type Colour string

const (
  Red   Colour = "red"
  Green Colour = "green"
  Blue  Colour = "blue"
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
```

  </TabItem>
</Tabs>

## Type aliases

A type alias is an alternate name for an existing type. It can be declared like so: 
