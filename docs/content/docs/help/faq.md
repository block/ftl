+++
title = "FAQ"
description = "Answers to frequently asked questions."
date = 2021-05-01T19:30:00+00:00
updated = 2021-05-01T19:30:00+00:00
draft = false
weight = 30
sort_by = "weight"
template = "docs/page.html"

[extra]
lead = "Answers to frequently asked questions."
toc = true
top = false
+++

## Why does FTL not allow external types directly?

Because of the nature of writing FTL verbs and data types, it's easy to think of it as just writing standard native code. Through that lens it is then somewhat surprising when FTL disallows the use of arbitrary external data types.

However, FTL types are not _just_ native types. FTL types are a more convenient method of writing an IDL such as [Protobufs](https://protobuf.dev/), [OpenAPI](https://www.openapis.org/) or [Thrift](https://thrift.apache.org/). With this in mind the constraint makes more sense. An IDL by its very nature must support a multitude of languages, so including an arbitrary type from a third party native library in one language may not be translatable to another language.

There are also secondary reasons, such as:

- Unclear ownership - in FTL a type must be owned by a single module. When importing a common third party type from multiple modules, which module owns that type?
- An external type must be representable in the FTL schema. The schema is then used to generate types for other modules, including those in other languages. For an external type, FTL could track the external library it belongs to to generate the "correct" code, but this would only be representable in a single language.
- External types often perform custom marshalling to/from JSON. This is not representable cross-language.
- Cleaner separation of abstraction layers - the ability to mix in abitrary external types is convenient, but can easily lead to mixing of concerns between internal and external data representations.

So what to do? See the [external types](https://block.github.io/ftl/docs/reference/externaltypes/) documentation
for how to work around this limitation.

## What is a "module"?

In its least abstract form, a module is a collection of verbs, and the resources (databases, queues, cron jobs, secrets, config, etc.) that those verbs rely on to operate. All resources are private to their owning module.

More abstractly, the separation of concerns between modules is largely subjective. You _can_ think of each module as largely analogous to a traditional service, so when asking where the division between modules is that could inform your decision. That said, the ease of deploying modules in FTL is designed to give you more flexibility in how you structure your code.
