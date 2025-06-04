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
import xyz.block.ftl.hotreload.CodeGenNotification;
import xyz.block.ftl.schema.v1.Data;
import xyz.block.ftl.schema.v1.Enum;
import xyz.block.ftl.schema.v1.EnumVariant;
import xyz.block.ftl.schema.v1.MetadataSQLColumn;
import xyz.block.ftl.schema.v1.MetadataSQLQuery;
import xyz.block.ftl.schema.v1.Module;
import xyz.block.ftl.schema.v1.Ref;
import xyz.block.ftl.schema.v1.Topic;
import xyz.block.ftl.schema.v1.Type;
import xyz.block.ftl.schema.v1.TypeAlias;
import xyz.block.ftl.schema.v1.Verb;
import xyz.block.ftl.schema.v1.Visibility;

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
        List<Path> schemaFiles = new ArrayList<>();
        log.debug("Generating JVM clients, data, enums from schema");
        Path generatedDir = context.inputDir().resolve("generated");
        if (!Files.isDirectory(context.inputDir())) {
            return false;
        }
        try {
            List<Module> modules = new ArrayList<>();
            Map<DeclRef, Type> typeAliasMap = new HashMap<>();
            Map<DeclRef, String> nativeTypeAliasMap = new HashMap<>();
            Map<DeclRef, List<EnumInfo>> enumVariantInfoMap = new HashMap<>();
            Map<String, PackageOutput> packageOutputMap = new HashMap<>();
            try (Stream<Path> pathStream = Files.list(context.inputDir())) {
                for (var file : pathStream.toList()) {
                    String fileName = file.getFileName().toString();
                    if (!fileName.endsWith(".pb")) {
                        continue;
                    }
                    schemaFiles.add(file);
                    Module module;
                    try {
                        module = Module.parseFrom(Files.readAllBytes(file));
                    } catch (Exception e) {
                        throw new CodeGenException("Failed to parse " + file, e);
                    }
                    String packageName = PACKAGE_PREFIX + module.getName();
                    var output = packageOutputMap.computeIfAbsent(module.getName(),
                            (k) -> new PackageOutput(context.outDir(), packageName));
                    for (var decl : module.getDeclsList()) {
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
                                                    output);
                                            handled = true;
                                            break;
                                        }
                                    }
                                }
                            }
                            if (!handled) {
                                generateTypeAliasMapper(module.getName(), data, packageName, Optional.empty(),
                                        output);
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
                    var output = packageOutputMap.computeIfAbsent(module.getName(),
                            (k) -> new PackageOutput(context.outDir(), packageName));
                    for (var decl : module.getDeclsList()) {
                        if (decl.hasVerb()) {
                            var verb = decl.getVerb();
                            if (verb.getVisibility() == Visibility.VISIBILITY_SCOPE_NONE) {
                                continue;
                            }

                            log.debugf("Generating verb %s", verb.getName());
                            generateVerb(module, verb, packageName, typeAliasMap, nativeTypeAliasMap, output);
                        } else if (decl.hasData()) {
                            var data = decl.getData();
                            if (data.getVisibility() == Visibility.VISIBILITY_SCOPE_NONE) {
                                continue;
                            }
                            log.debugf("Generating data %s", data.getName());
                            generateDataObject(module, data, packageName, typeAliasMap, nativeTypeAliasMap, enumVariantInfoMap,
                                    output);

                        } else if (decl.hasEnum()) {
                            var data = decl.getEnum();
                            if (data.getVisibility() == Visibility.VISIBILITY_SCOPE_NONE) {
                                continue;
                            }
                            log.debugf("Generating enum %s", data.getName());
                            generateEnum(module, data, packageName, typeAliasMap, nativeTypeAliasMap, enumVariantInfoMap,
                                    output);
                        } else if (decl.hasTopic()) {
                            var data = decl.getTopic();
                            if (data.getVisibility() == Visibility.VISIBILITY_SCOPE_NONE) {
                                continue;
                            }
                            log.debugf("Generating topic %s", data.getName());
                            generateTopicConsumer(module, data, packageName, typeAliasMap, nativeTypeAliasMap,
                                    output);
                        }
                    }
                }

                schemaFiles.addAll(writeGeneratedClients(context, packageOutputMap, generatedDir));

            } catch (Exception e) {
                throw new CodeGenException(e);
            }
            for (var e : packageOutputMap.entrySet()) {
                e.getValue().close();
            }
            return true;
        } finally {
            CodeGenNotification.updateLastModified(List.of(generatedDir, context.inputDir()), schemaFiles);
        }
    }

    protected abstract void generateTypeAliasMapper(String module, TypeAlias typeAlias, String packageName,
            Optional<String> nativeTypeAlias, PackageOutput outputDir) throws IOException;

    protected abstract void generateTopicConsumer(Module module, Topic data, String packageName,
            Map<DeclRef, Type> typeAliasMap, Map<DeclRef, String> nativeTypeAliasMap, PackageOutput outputDir)
            throws IOException;

    protected abstract void generateEnum(Module module, Enum data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Map<DeclRef, List<EnumInfo>> enumVariantInfoMap, PackageOutput outputDir)
            throws IOException;

    protected abstract void generateDataObject(Module module, Data data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Map<DeclRef, List<EnumInfo>> enumVariantInfoMap, PackageOutput outputDir)
            throws IOException;

    protected abstract void generateVerb(Module module, Verb verb, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, PackageOutput outputDir) throws IOException;

    protected abstract void generateSQLQueryVerb(Module module, Verb verb, String dbName, MetadataSQLQuery queryMetadata,
            String packageName, PackageOutput outputDir)
            throws IOException;

    private List<Path> writeGeneratedClients(CodeGenContext context, Map<String, PackageOutput> packageOutputMap,
            Path generatedDir)
            throws CodeGenException {
        List<Path> schemaFiles = new ArrayList<>();
        if (!Files.exists(generatedDir)) {
            log.debug("No generated clients found");
            return List.of();
        }
        try (Stream<Path> pathStream = Files.list(generatedDir)) {
            for (var file : pathStream.toList()) {
                String fileName = file.getFileName().toString();
                if (!fileName.endsWith(".pb")) {
                    continue;
                }
                Module module;
                try {
                    module = Module.parseFrom(Files.readAllBytes(file));
                } catch (Exception e) {
                    throw new CodeGenException("Failed to parse generated schema file " + file, e);
                }
                schemaFiles.add(file);

                String packageName = PACKAGE_PREFIX + module.getName();
                var output = packageOutputMap.computeIfAbsent(module.getName(),
                        (k) -> new PackageOutput(context.outDir(), packageName));
                for (var decl : module.getDeclsList()) {
                    if (decl.hasVerb()) {
                        log.debugf("Generating SQL verb %s", decl.getVerb().getName());
                        var verb = decl.getVerb();
                        var queryMetadata = verb.getMetadataList().stream().filter(md -> md.hasSqlQuery()).findFirst()
                                .map(md -> md.getSqlQuery()).orElse(null);
                        var dbCallMetadata = verb.getMetadataList().stream().filter(md -> md.hasDatabases()).findFirst()
                                .map(md -> md.getDatabases()).orElse(null);
                        var dbCalls = dbCallMetadata != null ? dbCallMetadata.getUsesList() : null;
                        if (queryMetadata != null) {
                            if (dbCalls == null) {
                                throw new CodeGenException("SQL query verb " + verb.getName() + " has no database calls");
                            }
                            if (dbCalls.size() != 1) {
                                throw new CodeGenException("SQL query verb " + verb.getName() + " has " + dbCalls.size()
                                        + " database calls, but must have exactly one");
                            }
                            var dbName = dbCalls.get(0).getName();
                            generateSQLQueryVerb(module, verb, dbName, queryMetadata, packageName, output);
                        }
                    } else if (decl.hasData()) {
                        log.debugf("Generating SQL verb data %s", decl.getVerb().getName());
                        var data = decl.getData();
                        generateDataObject(module, data, packageName, new HashMap<>(), new HashMap<>(), new HashMap<>(),
                                output);
                    }
                }
            }
        } catch (IOException e) {
            throw new CodeGenException(e);
        }
        return schemaFiles;
    }

    protected List<SQLColumnField> getOrderedSQLFields(Module module, Type type) {
        if (type.hasArray()) {
            return getOrderedSQLFields(module, type.getArray().getElement());
        } else if (type.hasMap()) {
            return getOrderedSQLFields(module, type.getMap().getValue());
        } else if (type.hasOptional()) {
            return getOrderedSQLFields(module, type.getOptional().getType());
        }

        Ref ref = type.getRef();
        if (ref == null) {
            return List.of();
        }
        Data data = resolveDataDecl(module, ref);
        if (data == null) {
            return List.of();
        }
        List<SQLColumnField> fields = new ArrayList<>();
        for (var field : data.getFieldsList()) {
            MetadataSQLColumn fieldMd = field.getMetadataList().stream().findFirst().map(md -> md.getSqlColumn())
                    .orElse(null);
            if (fieldMd != null) {
                fields.add(new SQLColumnField(field.getName(), fieldMd));
            }
        }
        return fields;
    }

    private Data resolveDataDecl(Module module, Ref ref) {
        if (ref == null) {
            return null;
        }
        return module.getDeclsList().stream().filter(d -> d.hasData() && d.getData().getName().equals(ref.getName()))
                .findFirst().map(d -> d.getData()).orElse(null);
    }

    @Override
    public boolean shouldRun(Path sourceDir, Config config) {
        return true;
    }

    public record SQLColumnField(String name, MetadataSQLColumn metadata) {
    }

    public record DeclRef(String module, String name) {
    }

    public record EnumInfo(String interfaceType, EnumVariant variant, List<EnumVariant> otherVariants) {
    }

    protected static String className(String in) {
        return Character.toUpperCase(in.charAt(0)) + in.substring(1);
    }

    protected static String camelToUpperSnake(String input) {
        StringBuilder snake = new StringBuilder();
        for (char c : input.toCharArray()) {
            if (Character.isUpperCase(c)) {
                if (snake.length() > 0) {
                    snake.append('_');
                }
                snake.append(c);
            } else {
                snake.append(Character.toUpperCase(c));
            }
        }
        return snake.toString();
    }
}
