package xyz.block.ftl.kotlin.compiler.plugin

import org.jetbrains.kotlin.GeneratedDeclarationKey
import org.jetbrains.kotlin.descriptors.ClassKind
import org.jetbrains.kotlin.descriptors.Visibilities
import org.jetbrains.kotlin.fir.FirSession
import org.jetbrains.kotlin.fir.declarations.FirDeclarationOrigin
import org.jetbrains.kotlin.fir.declarations.FirRegularClass
import org.jetbrains.kotlin.fir.declarations.builder.buildRegularClass
import org.jetbrains.kotlin.fir.declarations.utils.memberDeclarationNameOrNull
import org.jetbrains.kotlin.fir.extensions.ExperimentalTopLevelDeclarationsGenerationApi
import org.jetbrains.kotlin.fir.extensions.FirDeclarationGenerationExtension
import org.jetbrains.kotlin.fir.extensions.MemberGenerationContext
import org.jetbrains.kotlin.fir.extensions.NestedClassGenerationContext
import org.jetbrains.kotlin.fir.moduleData
import org.jetbrains.kotlin.fir.plugin.*
import org.jetbrains.kotlin.fir.resolve.shortName
import org.jetbrains.kotlin.fir.symbols.impl.*
import org.jetbrains.kotlin.fir.types.classId
import org.jetbrains.kotlin.fir.types.coneType
import org.jetbrains.kotlin.fir.types.impl.FirUserTypeRefImpl
import org.jetbrains.kotlin.name.*

private const val ftlClient = "FTLClient"

class FIRClientGenerationExtension(session: FirSession) : FirDeclarationGenerationExtension(session) {

  override fun generateNestedClassLikeDeclaration(
    owner: FirClassSymbol<*>,
    name: Name,
    context: NestedClassGenerationContext
  ): FirClassLikeSymbol<*>? {
    if (!name.asString().endsWith(ftlClient)) {
      return null
    }
    val klazz = createNestedClass(owner, name, Key(name.asString()), ClassKind.INTERFACE)

    return klazz.symbol
  }

  override fun getNestedClassifiersNames(
    classSymbol: FirClassSymbol<*>,
    context: NestedClassGenerationContext
  ): Set<Name> {
    println("FOOOO" + classSymbol.name.asString())
    if (classSymbol.name.asString().endsWith(ftlClient)) {
      return emptySet()
    }
    val ret = HashSet<Name>()
    ret.add(Name.identifier("Test$ftlClient"))
    for (i in classSymbol.declarationSymbols) {
      if (i.memberDeclarationNameOrNull == null) {
        continue
      }
      for (ann in i.annotations) {
        val userTypeRef = ann.annotationTypeRef as? FirUserTypeRefImpl
        if (userTypeRef != null) {
          // The name of the type (last part, e.g., "C" in A.B.C)
          val simpleName = userTypeRef.shortName
          // The full qualifier (e.g., "A.B" in A.B.C)
          val qualifier = userTypeRef.qualifier.joinToString(".") // "A.B" if exists
          if (simpleName.equals("Verb") && qualifier == "xyz.block.ftl") {
            ret.add(Name.identifier(classSymbol.name.asString() + "$" + i.memberDeclarationNameOrNull!! + ftlClient))
            break
          }
        } else {
          if (ann.annotationTypeRef.coneType.classId?.asSingleFqName() == FqName("xyz.block.ftl.Verb")) {
            ret.add(Name.identifier(classSymbol.name.asString() + "$" + i.memberDeclarationNameOrNull!! + ftlClient))
            break
          }
        }
      }
    }

    return ret
  }

  data class Key(val name : String) : GeneratedDeclarationKey() {

  }
}
