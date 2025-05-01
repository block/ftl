package xyz.block.ftl.kotlin.compiler.plugin

import org.jetbrains.kotlin.fir.extensions.FirExtensionRegistrar

class SimplePluginRegistrar : FirExtensionRegistrar() {
    override fun ExtensionRegistrarContext.configurePlugin() {
        +::FIRClientGenerationExtension
    }
}
