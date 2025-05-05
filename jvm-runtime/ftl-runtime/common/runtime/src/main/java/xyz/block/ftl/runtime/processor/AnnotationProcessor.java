package xyz.block.ftl.runtime.processor;

import java.io.BufferedWriter;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

import javax.annotation.processing.Completion;
import javax.annotation.processing.ProcessingEnvironment;
import javax.annotation.processing.Processor;
import javax.annotation.processing.RoundEnvironment;
import javax.lang.model.SourceVersion;
import javax.lang.model.element.*;
import javax.lang.model.type.ArrayType;
import javax.lang.model.type.DeclaredType;
import javax.lang.model.type.TypeKind;
import javax.lang.model.type.TypeMirror;
import javax.tools.Diagnostic;
import javax.tools.FileObject;
import javax.tools.StandardLocation;

import xyz.block.ftl.*;
import xyz.block.ftl.Enum;
import xyz.block.ftl.Export;
import xyz.block.ftl.Secret;
import xyz.block.ftl.TypeAlias;
import xyz.block.ftl.Verb;

/**
 * POC annotation processor for capturing JavaDoc, this needs a lot more work.
 */
public class AnnotationProcessor implements Processor {

    private static final Pattern REMOVE_LEADING_SPACE = Pattern.compile("^ ", Pattern.MULTILINE);
    private static final Pattern REMOVE_JAVADOC_TAGS = Pattern.compile(
            "^\\s*@(param|return|throws|exception|see|author)\\b[^\\n]*$\\n*",
            Pattern.MULTILINE);
    private ProcessingEnvironment processingEnv;

    final Map<String, String> saved = new HashMap<>();
    final Set<String> processedLocalVerbs = new HashSet<>();

    @Override
    public Set<String> getSupportedOptions() {
        return Set.of();
    }

    @Override
    public Set<String> getSupportedAnnotationTypes() {
        return Set.of(Verb.class.getName(), Enum.class.getName(), Export.class.getName(), TypeAlias.class.getName());
    }

    @Override
    public SourceVersion getSupportedSourceVersion() {
        return SourceVersion.latestSupported();
    }

    @Override
    public void init(ProcessingEnvironment processingEnv) {
        this.processingEnv = processingEnv;
    }

    @Override
    public boolean process(Set<? extends TypeElement> annotations, RoundEnvironment roundEnv) {
        //TODO: HTTP etc
        roundEnv.getElementsAnnotatedWithAny(Set.of(Verb.class, Enum.class, Data.class, Cron.class, TypeAlias.class))
                .forEach(element -> {
                    Optional<String> javadoc = getJavadoc(element);
                    javadoc.ifPresent(doc -> {
                        String strippedDownDoc = stripJavadocTags(doc);
                        String name;
                        if (element.getAnnotation(TypeAlias.class) != null) {
                            name = element.getAnnotation(TypeAlias.class).name();
                        } else {
                            name = element.getSimpleName().toString();
                        }
                        saved.put(name, strippedDownDoc);
                        if (element.getKind() == ElementKind.METHOD) {
                            var executableElement = (ExecutableElement) element;
                            executableElement.getParameters().forEach(param -> {
                                Config config = param.getAnnotation(Config.class);
                                if (config != null) {
                                    saved.put(config.value(), extractCommentForParam(doc, param));
                                }
                                Secret secret = param.getAnnotation(Secret.class);
                                if (secret != null) {
                                    saved.put(secret.value(), extractCommentForParam(doc, param));
                                }
                            });
                        }
                    });
                });

        roundEnv.getElementsAnnotatedWithAny(Set.of(Verb.class))
                .forEach(element -> {
                    if (element.getKind() == ElementKind.METHOD) {
                        var executableElement = (ExecutableElement) element;
                        String verbName = executableElement.getSimpleName().toString();
                        if (processedLocalVerbs.contains(verbName)) {
                            return;
                        }
                        String className = verbName;
                        className = Character.toUpperCase(className.charAt(0)) + className.substring(1);

                        try {
                            List<? extends VariableElement> parameters = executableElement.getParameters();
                            String paramName = "";
                            if (!parameters.isEmpty()) {
                                for (var i = 0; i < parameters.size(); i++) {
                                    var elem = parameters.get(i);
                                    if (elem.getAnnotation(Config.class) != null || elem.getAnnotation(Secret.class) != null) {
                                        continue;
                                    }
                                    boolean ok = true;
                                    if (elem.asType().getKind() == TypeKind.DECLARED) {
                                        DeclaredType declaredType = (DeclaredType) elem.asType();
                                        if (declaredType.toString().contains("error.NonExistentClass")) {
                                            continue;
                                        }
                                        if (declaredType.asElement().getAnnotation(Topic.class) != null) {
                                            ok = false;
                                        }
                                        if (ok) {
                                            for (var meth : declaredType.asElement().getEnclosedElements()) {
                                                if (meth.getAnnotation(VerbClient.class) != null) {
                                                    ok = false;
                                                    break;
                                                }
                                            }
                                        }
                                    }
                                    if (!ok) {
                                        continue;
                                    }
                                    paramName = typeMirrorToString(elem.asType());
                                }
                            }
                            var returnType = typeMirrorToString(executableElement.getReturnType());
                            String iface;
                            if (returnType.equals("void")) {
                                if (paramName.isEmpty()) {
                                    iface = "xyz.block.ftl.EmptyVerb";
                                } else {
                                    iface = "xyz.block.ftl.SinkVerb<" + paramName + ">";
                                }
                            } else {
                                if (paramName.isEmpty()) {
                                    iface = "xyz.block.ftl.SourceVerb<" + returnType + ">";
                                } else {
                                    iface = "xyz.block.ftl.FunctionVerb<" + paramName + "," + returnType + ">";
                                }
                            }

                            if (!paramName.isEmpty()) {
                                paramName = "val: " + paramName;
                            }
                            String kotlinDir = processingEnv.getOptions().get("kapt.kotlin.generated");

                            if (kotlinDir != null) {
                                var path = Paths.get(kotlinDir, "client").normalize();
                                Files.createDirectories(path);

                                var file = path.resolve(className + "Client.kt");
                                var template = """
                                        package client;

                                        import xyz.block.ftl.VerbClient

                                        @VerbClient(name="$VERBNAME")
                                        public interface $CLASSNAMEClient: $IFACE {
                                            override fun call($PARAM):$RETURN
                                        }
                                        """;
                                template = template.replace("$CLASSNAME", className);
                                template = template.replace("$RETURN", returnType);
                                template = template.replace("$VERBNAME", verbName);
                                template = template.replace("$PARAM", paramName);
                                template = template.replace("$IFACE", iface);
                                Files.writeString(file, template);
                            } else {

                                var file = processingEnv.getFiler().createSourceFile("client." + className + "Client");
                                var template = """
                                        package client;

                                        import xyz.block.ftl.VerbClient;

                                        @VerbClient(name="$VERBNAME")
                                        public interface $CLASSNAMEClient extends $IFACE {
                                            $RETURN call($PARAM);
                                        }
                                        """;
                                template = template.replace("$CLASSNAME", className);
                                template = template.replace("$RETURN", returnType);
                                template = template.replace("$VERBNAME", verbName);
                                template = template.replace("$PARAM", paramName);
                                template = template.replace("$IFACE", iface);
                                try (var writer = file.openWriter()) {
                                    writer.append(template);
                                }
                            }
                            processedLocalVerbs.add(verbName);
                        } catch (IgnoreException e) {
                            this.processingEnv.getMessager().printWarning(e.getMessage(), element);
                        } catch (IOException e) {
                            throw new RuntimeException(e);
                        }

                    }
                });

        if (roundEnv.processingOver()) {
            write("META-INF/ftl-verbs.txt", saved.entrySet().stream().map(
                    e -> e.getKey() + "=" + Base64.getEncoder().encodeToString(e.getValue().getBytes(StandardCharsets.UTF_8)))
                    .collect(Collectors.toSet()));
        }
        return false;
    }

    private static String typeMirrorToString(TypeMirror typeMirror) throws IgnoreException {
        switch (typeMirror.getKind()) {
            case BOOLEAN:
                return "boolean";
            case BYTE:
                return "byte";
            case SHORT:
                return "short";
            case INT:
                return "int";
            case LONG:
                return "long";
            case FLOAT:
                return "float";
            case DOUBLE:
                return "double";
            case CHAR:
                return "char";
            case ARRAY:
                ArrayType arrayType = (ArrayType) typeMirror;
                return typeMirrorToString(arrayType.getComponentType()) + "[]";
            case DECLARED:
                DeclaredType type = (DeclaredType) typeMirror;
                Element element = type.asElement();
                Element enclosing = element;
                while (enclosing.getKind() != ElementKind.PACKAGE) {
                    enclosing = enclosing.getEnclosingElement();
                }
                PackageElement packageElement = (PackageElement) enclosing;
                return packageElement.getQualifiedName().toString() + "." + element.getSimpleName().toString();
            case VOID:
                return "void";
            case ERROR:
                throw new IgnoreException("Unknown type " + typeMirror.toString());
            default:
                throw new IllegalArgumentException("Unsupported type: " + typeMirror.toString());
        }
    }

    /**
     * This method uses the annotation processor Filer API and we shouldn't use a Path as paths containing \ are not supported.
     */
    public void write(String filePath, Set<String> set) {
        if (set.isEmpty()) {
            return;
        }
        try {
            final FileObject listResource = processingEnv.getFiler().createResource(StandardLocation.CLASS_OUTPUT, "",
                    filePath.toString());

            try (BufferedWriter writer = new BufferedWriter(
                    new OutputStreamWriter(listResource.openOutputStream(), StandardCharsets.UTF_8))) {
                for (String className : set) {
                    writer.write(className);
                    writer.newLine();
                }
            }
        } catch (IOException e) {
            processingEnv.getMessager().printMessage(Diagnostic.Kind.ERROR, "Failed to write " + filePath + ": " + e);
            return;
        }
    }

    @Override
    public Iterable<? extends Completion> getCompletions(Element element, AnnotationMirror annotation, ExecutableElement member,
            String userText) {
        return null;
    }

    public Optional<String> getJavadoc(Element e) {
        String docComment = processingEnv.getElementUtils().getDocComment(e);

        if (docComment == null || docComment.isBlank()) {
            return Optional.empty();
        }

        // javax.lang.model keeps the leading space after the "*" so we need to remove it.

        return Optional.of(REMOVE_LEADING_SPACE.matcher(docComment)
                .replaceAll("")
                .trim());
    }

    public String stripJavadocTags(String doc) {
        // TODO extract JavaDoc tags to a rich markdown model supported by schema
        return REMOVE_JAVADOC_TAGS.matcher(doc).replaceAll("");
    }

    /**
     * Read the @param tag in a JavaDoc comment to extract Config and Secret comments
     */
    private String extractCommentForParam(String doc, VariableElement param) {
        String variableName = param.getSimpleName().toString();
        int startIdx = doc.indexOf("@param " + variableName + " ");
        if (startIdx != -1) {
            int endIndex = doc.indexOf("\n", startIdx);
            if (endIndex == -1) {
                endIndex = doc.length();
            }
            return doc.substring(startIdx + variableName.length() + 8, endIndex);
        }
        return null;
    }

    private static class IgnoreException extends Exception {
        public IgnoreException(String message) {
            super(message);
        }

    }
}
