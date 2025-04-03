package xyz.block.ftl.test

import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import xyz.block.ftl.Egress

class Verbs {


  @Export
  @Verb
  fun valueEnumVerb(color: ColorWrapper): ColorWrapper {
    return color
  }

  @Export
  @Verb
  fun stringEnumVerb(shape: ShapeWrapper): ShapeWrapper {
    return shape
  }

  @Export
  @Verb
  fun typeEnumVerb(animal: AnimalWrapper): AnimalWrapper {
    return animal
  }

  @Export
  @Verb
  fun egress(@Egress("url") url: String): String {
    return url
  }
}
