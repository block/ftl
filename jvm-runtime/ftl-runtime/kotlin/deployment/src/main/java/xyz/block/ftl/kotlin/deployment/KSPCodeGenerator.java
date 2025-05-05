package xyz.block.ftl.kotlin.deployment;

import java.nio.file.Path;
import java.util.List;
import java.util.Map;

import org.eclipse.microprofile.config.Config;

import com.google.devtools.ksp.impl.CommandLineKSPLogger;
import com.google.devtools.ksp.impl.KotlinSymbolProcessing;
import com.google.devtools.ksp.processing.KSPJvmConfig;

import io.quarkus.bootstrap.model.ApplicationModel;
import io.quarkus.bootstrap.prebuild.CodeGenException;
import io.quarkus.deployment.CodeGenContext;
import io.quarkus.deployment.CodeGenProvider;

public class KSPCodeGenerator implements CodeGenProvider {

    @Override
    public String providerId() {
        return "ftl-ksp";
    }

    @Override
    public String[] inputExtensions() {
        return new String[] { "kt" };
    }

    @Override
    public String inputDirectory() {
        return "kotlin";
    }

    @Override
    public void init(ApplicationModel model, Map<String, String> properties) {
        CodeGenProvider.super.init(model, properties);
    }

    @Override
    public boolean trigger(CodeGenContext context) throws CodeGenException {
        KSPJvmConfig.Builder builder = new KSPJvmConfig.Builder();
        builder.javaOutputDir = context.outDir().toFile();
        builder.jvmTarget = "17";
        builder.moduleName = "";
        builder.sourceRoots = List.of(context.inputDir().toFile());
        builder.projectBaseDir = context.inputDir().toFile();
        builder.cachesDir = context.outDir().toFile();
        builder.languageVersion = "2";
        builder.outputBaseDir = context.outDir().toFile();
        builder.resourceOutputDir = context.outDir().toFile();
        builder.classOutputDir = context.outDir().toFile();
        builder.kotlinOutputDir = context.outDir().toFile();
        builder.apiVersion = "2";

        KotlinSymbolProcessing ksp = new KotlinSymbolProcessing(builder.build(), List.of(),
                new CommandLineKSPLogger());
        ksp.execute();
        return true;
    }

    @Override
    public boolean shouldRun(Path sourceDir, Config config) {
        return true;
    }
}
