package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Stream;

import org.eclipse.microprofile.config.Config;
import org.jboss.logging.Logger;

import io.quarkus.bootstrap.prebuild.CodeGenException;
import io.quarkus.deployment.CodeGenContext;
import io.quarkus.deployment.CodeGenProvider;
import xyz.block.ftl.schema.v1.Data;
import xyz.block.ftl.schema.v1.Enum;
import xyz.block.ftl.schema.v1.EnumVariant;
import xyz.block.ftl.schema.v1.Module;
import xyz.block.ftl.schema.v1.Topic;
import xyz.block.ftl.schema.v1.Type;
import xyz.block.ftl.schema.v1.TypeAlias;
import xyz.block.ftl.schema.v1.Verb;

public abstract class JVMCodeGenerator implements CodeGenProvider {

    public static final String PACKAGE_PREFIX = "ftl.";
    public static final String TYPE_MAPPER = "TypeAliasMapper";
    private static final Logger log = Logger.getLogger(JVMCodeGenerator.class);

    @Override
    public String providerId() {
        return "ftl-clients";
    }

    @Override
    public String inputDirectory() {
        return "ftl-module-schema";
    }

    @Override
    public boolean trigger(CodeGenContext context) throws CodeGenException {
        log.info("Generating JVM clients, data, enums from schema");
        if (!Files.isDirectory(context.inputDir())) {
            return false;
        }
        List<Module> modules = new ArrayList<>();
        Map<DeclRef, Type> typeAliasMap = new HashMap<>();
        Map<DeclRef, String> nativeTypeAliasMap = new HashMap<>();
        Map<DeclRef, List<EnumInfo>> enumVariantInfoMap = new HashMap<>();
        try (Stream<Path> pathStream = Files.list(context.inputDir())) {
            for (var file : pathStream.toList()) {
                String fileName = file.getFileName().toString();
                if (!fileName.endsWith(".pb")) {
                    continue;
                }
                Module module;
                try {
                    module = Module.parseFrom(Files.readAllBytes(file));
                } catch (Exception e) {
                    throw new CodeGenException("Failed to parse " + file, e);
                }
                for (var decl : module.getDeclsList()) {
                    String packageName = PACKAGE_PREFIX + module.getName();
                    if (decl.hasTypeAlias()) {
                        var data = decl.getTypeAlias();
                        boolean handled = false;
                        for (var md : data.getMetadataList()) {
                            if (md.hasTypeMap()) {
                                String runtime = md.getTypeMap().getRuntime();
                                if (runtime.equals("kotlin") || runtime.equals("java")) {
                                    String nativeName = md.getTypeMap().getNativeName();
                                    var existing = getClass().getClassLoader()
                                            .getResource(nativeName.replace(".", "/") + ".class");
                                    if (existing != null) {
                                        nativeTypeAliasMap.put(new DeclRef(module.getName(), data.getName()),
                                                nativeName);
                                        generateTypeAliasMapper(module.getName(), data, packageName,
                                                Optional.of(nativeName),
                                                context.outDir());
                                        handled = true;
                                        break;
                                    }
                                }
                            }
                        }
                        if (!handled) {
                            generateTypeAliasMapper(module.getName(), data, packageName, Optional.empty(),
                                    context.outDir());
                            typeAliasMap.put(new DeclRef(module.getName(), data.getName()), data.getType());
                        }

                    }
                }
                modules.add(module);
            }
        } catch (IOException e) {
            throw new CodeGenException(e);
        }
        try {
            for (var module : modules) {
                String packageName = PACKAGE_PREFIX + module.getName();
                for (var decl : module.getDeclsList()) {
                    if (decl.hasVerb()) {
                        var verb = decl.getVerb();
                        if (!verb.getExport()) {
                            continue;
                        }
                        generateVerb(module, verb, packageName, typeAliasMap, nativeTypeAliasMap, context.outDir());
                    } else if (decl.hasData()) {
                        var data = decl.getData();
                        if (!data.getExport()) {
                            continue;
                        }
                        generateDataObject(module, data, packageName, typeAliasMap, nativeTypeAliasMap, enumVariantInfoMap,
                                context.outDir());

                    } else if (decl.hasEnum()) {
                        var data = decl.getEnum();
                        if (!data.getExport()) {
                            continue;
                        }
                        generateEnum(module, data, packageName, typeAliasMap, nativeTypeAliasMap, enumVariantInfoMap,
                                context.outDir());
                    } else if (decl.hasTopic()) {
                        var data = decl.getTopic();
                        if (!data.getExport()) {
                            continue;
                        }
                        generateTopicConsumer(module, data, packageName, typeAliasMap, nativeTypeAliasMap,
                                context.outDir());
                    }
                }
            }

        } catch (Exception e) {
            throw new CodeGenException(e);
        }
        return true;
    }

    protected abstract void generateTypeAliasMapper(String module, TypeAlias typeAlias, String packageName,
            Optional<String> nativeTypeAlias, Path outputDir) throws IOException;

    protected abstract void generateTopicConsumer(Module module, Topic data, String packageName,
            Map<DeclRef, Type> typeAliasMap, Map<DeclRef, String> nativeTypeAliasMap, Path outputDir) throws IOException;

    protected abstract void generateEnum(Module module, Enum data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Map<DeclRef, List<EnumInfo>> enumVariantInfoMap, Path outputDir)
            throws IOException;

    protected abstract void generateDataObject(Module module, Data data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Map<DeclRef, List<EnumInfo>> enumVariantInfoMap, Path outputDir)
            throws IOException;

    protected abstract void generateVerb(Module module, Verb verb, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Path outputDir) throws IOException;

    @Override
    public boolean shouldRun(Path sourceDir, Config config) {
        return true;
    }

    public record DeclRef(String module, String name) {
    }

    public record EnumInfo(String interfaceType, EnumVariant variant, List<EnumVariant> otherVariants) {
    }

    protected static String className(String in) {
        return Character.toUpperCase(in.charAt(0)) + in.substring(1);
    }

}
