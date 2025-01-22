+++
title = "Types"
description = "Declaring and using Types"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 30
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

FTL supports the following types: `Int` (64-bit), `Float` (64-bit), `String`, `Bytes` (a byte array), `Bool`, `Time`,
`Any` (a dynamic type), `Unit` (similar to "void"), arrays, maps, data structures, and constant enumerations. Each FTL
type is mapped to a corresponding language-specific type. For example in Go `Float` is represented as `float64`, `Time`
is represented by `time.Time`, and so on.

User-defined types referenced by a verb will be automatically exported as FTL types.

| FTL        | Go                           | Kotlin                       | Java (Optional)              |
|:-----------|:-----------------------------|:-----------------------------|:-----------------------------|
| `Int`      | `int`                        | `Long`                       | `long (Long)`                |
| `Float`    | `float64`                    | `Double`                     | `double (Double)`            |
| `String`   | `string`                     | `String`                     | `String`                     |
| `Bytes`    | `[]byte`                     | `ByteArray`                  | `byte[]`                     |
| `Bool`     | `bool`                       | `Boolean`                    | `boolean (Boolean)`          |
| `Time`     | `time.Time`                  | `ZonedDateTime`              | `ZonedDateTime️`             |
| `Any`      | [external](../externaltypes) | [external](../externaltypes) | [external](../externaltypes) |
| `Unit`     |                              |                              |                              |
| `Map<K,V>` | `map[K]V`                    | `Map<K,V>`                   | `Map<K,V>`                   |
| `Array<T>` | `[]T`                        | `List<T>`                    | `List<T>`                    |

Go types can be declared as optional using the `ftl.Option[T]` type. Kotlin uses the built in nullable `?` suffix to
denote optional.
For Java if an explicit optional types is not called out in the table about the `@Nullable` annotation can be used to
denote optional types.

## Data structures

FTL supports user-defined data structures, declared using the idiomatic syntax of the target language.

{% code_selector() %}

<!-- go -->

```go
type Person struct {
  Name string
  Age  int
}
```

<!-- kotlin -->

```kotlin
data class Person(
  val name: String,
  val age: Int
)
```

<!-- java -->

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

{% end %}

## Generics

FTL has first-class support for generics, declared using the idiomatic syntax of the target language.

{% code_selector() %}

<!-- go -->

```go
type Pair[T, U] struct {
  First  T
  Second U
}
```

<!-- kotlin -->

```kotlin
data class Pair<T, U>(
  val first: T,
  val second: U
)
```

<!-- java -->

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

{% end %}

## Sum types

[Sum types](https://en.wikipedia.org/wiki/Tagged_union) are supported by FTL's type system.

{% code_selector() %}
<!-- go -->

Sum types aren't directly supported by Go, however they can be approximated with the use of [sealed interfaces](https://blog.chewxy.com/2018/03/18/golang-interfaces/):

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}

type Dog struct {}
func (Dog) animal() {}
```

<!-- kotlin -->

Sum types aren't directly supported by Kotlin, however they can be approximated with the use of [sealed interfaces](https://kotlinlang.org/docs/sealed-classes.html):

```kotlin
@Enum
sealed interface Animal

@EnumHolder
class Cat() : Animal

@EnumHolder
class Dog() : Animal
```

<!-- java -->

> TODO

{% end %}

## Enumerations

A value enum is an enumerated set of string or integer values.

{% code_selector() %}
<!-- go -->

```go
//ftl:enum
type Colour string

const (
  Red   Colour = "red"
  Green Colour = "green"
  Blue  Colour = "blue"
)
```

<!-- kotlin -->

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

<!-- java -->

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

{% end %}

## Type aliases

A type alias is an alternate name for an existing type. It can be declared like so:

{% code_selector() %}

<!-- go -->

```go
//ftl:typealias
type Alias Target
```
or
```go
//ftl:typealias
type Alias = Target
```

eg.

```go
//ftl:typealias
type UserID string

//ftl:typealias
type UserToken = string
```

<!-- kotlin -->

> TODO

<!-- java -->

> TODO

{% end %}


## Optional types

FTL supports optional types where the underlying value can be present or absent.

{% code_selector() %}

<!-- go -->

In Go optional types can be declared via `ftl.Option[T]`. These types are provided by the `ftl` runtimes. For example,
the following FTL type declaration in go, will provide an optional string type "Name":

```go
type EchoResponse struct {
	Name ftl.Option[string]
}
```

The value of this type can be set to `Some` or `None`:

```go
resp := EchoResponse{
  Name: ftl.Some("John"),
}

resp := EchoResponse{
  Name: ftl.None(),
}
```

The value of the optional type can be accessed using `Get`, `MustGet`, or `Default` methods:

```go
// Get returns the value and a boolean indicating if the Option contains a value.
if value, ok := resp.Name.Get(); ok {
  resp.Name = ftl.Some(value)
}

// MustGet returns the value or panics if the Option is None.
value := resp.Name.MustGet()

// Default returns the value or a default value if the Option is None.
value := resp.Name.Default("default")
```

<!-- kotlin -->

Kotlin has built-in support for optional types:

```kotlin
data class EchoResponse(
  val name: String?
)
```

<!-- java -->

> TODO: Java optional types

{% end %}

## Unit "void" type

The [`Unit`](https://en.wikipedia.org/wiki/Unit_type) type is used to indicate that a value is not present. It is similar to C's `void` type.

For verbs, omitting the return or request type is equivalent to specifying `Unit`.

A function of the form `F(I) -> O` is known as a "verb", a function of the form `F(R)` is known as a "sink", a verb of
the form `F() -> R` is known as a "source", and a verb of the form `F()` is known as an "empty" verb.

{% code_selector() %}

<!-- go -->

```go
//ftl:verb
func Hello(ctx context.Context, req ftl.Unit) (string, error) {
  return "Hello, World!", nil
}
```

This is equivalent to:

```go
//ftl:verb
func Hello(ctx context.Context) (string, error) {
  return "Hello, World!", nil
}
```

<!-- kotlin -->

```kotlin
@Verb
fun hello(ctx: Context): String {
  return "Hello, World!"
}
```

<!-- java -->

```java
public class Hello {
  @Verb
  public String hello(Context ctx) {
    return "Hello, World!";
  }
}
```

{% end %}

## Builtin types

FTL provides a set of builtin types that are automatically available in all FTL runtimes. These types are:

- `builtin.HttpRequest<Body, PathParams, QueryParams>` - Represents an HTTP request with a body of type `Body`, path parameter type of `PathParams` and a query parameter type of `QueryParams`.
- `builtin.HttpResponse<Body, Error>` - Represents an HTTP response with a body of type `Body` and an error of type `Error`.
- `builtin.Empty` - Represents an empty type. This equates to an empty structure `{}`.
- `builtin.CatchRequest` - Represents a request structure for catch verbs.
