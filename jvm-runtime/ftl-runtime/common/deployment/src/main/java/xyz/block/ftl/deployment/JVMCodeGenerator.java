package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import org.eclipse.microprofile.config.Config;
import org.jboss.logging.Logger;
import org.tomlj.Toml;
import org.tomlj.TomlParseResult;

import io.quarkus.bootstrap.prebuild.CodeGenException;
import io.quarkus.deployment.CodeGenContext;
import io.quarkus.deployment.CodeGenProvider;
import xyz.block.ftl.schema.v1.Data;
import xyz.block.ftl.schema.v1.Decl;
import xyz.block.ftl.schema.v1.Enum;
import xyz.block.ftl.schema.v1.EnumVariant;
import xyz.block.ftl.schema.v1.Module;
import xyz.block.ftl.schema.v1.Ref;
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
    public String[] inputExtensions() {
        return new String[] { "pb" };
    }

    @Override
    public boolean trigger(CodeGenContext context) throws CodeGenException {
        log.info("Generating JVM clients, data, enums from schema");
        if (!Files.isDirectory(context.inputDir())) {
            log.warnf("Input directory %s does not exist", context.inputDir());
            return false;
        }
        List<Module> modules = new ArrayList<>();
        Map<DeclRef, Type> typeAliasMap = new HashMap<>();
        Map<DeclRef, String> nativeTypeAliasMap = new HashMap<>();
        Map<DeclRef, List<EnumInfo>> enumVariantInfoMap = new HashMap<>();

        log.warnf("Input directory: %s", context.inputDir());
        String currentModuleName = getCurrentModuleName(context.inputDir());
        log.warnf("Current module: %s", currentModuleName);

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
                log.warnf("Parsed module: %s", module.getName());
                for (var decl : module.getDeclsList()) {
                    log.warnf("Decl: %s", decl);
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
                log.warnf("Generating module: %s", module.getName());
                String packageName = PACKAGE_PREFIX + module.getName();
                for (var decl : module.getDeclsList()) {
                    log.warnf("Generating decl: %s", decl);
                    if (decl.hasVerb()) {
                        var verb = decl.getVerb();
                        boolean isLocal = currentModuleName != null && module.getName().equals(currentModuleName);
                        if (!verb.getExport() && !isLocal) {
                            continue;
                        }
                        //Path targetDir = isLocal && isQueryVerb(verb) ? getQueryDir(context.outDir()) : context.outDir();
                        log.warnf("Generating verb %s in %s", verb.getName(), context.outDir());
                        generateVerb(module, verb, packageName, typeAliasMap, nativeTypeAliasMap, context.outDir(), isLocal, () -> resolveDataDecl(module, verb.getRequest()));
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

    protected String getCurrentModuleName(Path source) {
        while (source != null) {
            Path toml = source.resolve("ftl.toml");
            if (Files.exists(toml)) {
                try {
                    TomlParseResult result = Toml.parse(toml);
                    if (result.hasErrors()) {
                        log.errorf("Failed to parse %s: %s", toml,
                                result.errors().stream().map(Objects::toString).collect(Collectors.joining(", ")));
                        return null;
                    }

                    String value = result.getString("module");
                    if (value != null) {
                        return value;
                    } else {
                        log.errorf("module name not found in %s", toml);
                    }
                } catch (IOException e) {
                    log.errorf("Failed to read %s: %s", toml, e.getMessage());
                }
            }
            source = source.getParent();
        }
        return null;
    }

    // protected Path getQueryDir(Path outDir) throws CodeGenException {
    //     Path queriesDir = outDir.getParent().resolve("ftl-queries");
    //     try {
    //         if (!Files.exists(queriesDir)) {
    //             Files.createDirectories(queriesDir);
    //             log.debugf("Created query directory at %s", queriesDir);
    //         }
    //         return queriesDir;
    //     } catch (IOException e) {
    //         throw new CodeGenException("Failed to create ftl-queries directory", e);
    //     }
    // }

    protected boolean isQueryVerb(Verb verb) {
        return verb.getMetadataList().stream().anyMatch(md -> md.hasSqlQuery());
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
            Map<DeclRef, String> nativeTypeAliasMap, Path outputDir, boolean isLocal) throws IOException;
    
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
