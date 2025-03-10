package xyz.block.ftl.runtime.processor;

import java.io.BufferedWriter;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.lang.reflect.Modifier;
import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

import javax.annotation.processing.Completion;
import javax.annotation.processing.ProcessingEnvironment;
import javax.annotation.processing.Processor;
import javax.annotation.processing.RoundEnvironment;
import javax.lang.model.SourceVersion;
import javax.lang.model.element.AnnotationMirror;
import javax.lang.model.element.Element;
import javax.lang.model.element.ElementKind;
import javax.lang.model.element.ExecutableElement;
import javax.lang.model.element.TypeElement;
import javax.lang.model.element.VariableElement;
import javax.tools.Diagnostic;
import javax.tools.FileObject;
import javax.tools.StandardLocation;

import io.quarkus.gizmo.ClassCreator;
import io.quarkus.gizmo.ClassOutput;
import io.quarkus.gizmo.MethodCreator;
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
                        String className = executableElement.getSimpleName().toString();
                        className = Character.toUpperCase(className.charAt(0)) + className.substring(1);
                        ClassOutput output = new ClassOutput() {
                            @Override
                            public void write(String name, byte[] data) {
                                try {
                                    var file = processingEnv.getFiler().createClassFile(name.replace('/', '.'));
                                    try (var out = file.openOutputStream()) {
                                        out.write(data);
                                    }
                                } catch (IOException e) {
                                    throw new RuntimeException(e);
                                }
                            }
                        };
                        try (ClassCreator creator = ClassCreator.interfaceBuilder()
                                .classOutput(output)
                                .className("client." + className + "Client")
                                .build()) {
                            List<? extends VariableElement> parameters = executableElement.getParameters();
                            String[] paramNames;
                            if (parameters.isEmpty()) {
                                paramNames = new String[0];
                            } else {
                                paramNames = new String[1];
                                for (var i = 0; i < parameters.size(); i++) {
                                    var elem = parameters.get(i);
                                    if (elem.getAnnotation(Config.class) != null || elem.getAnnotation(Secret.class) != null) {
                                        continue;
                                    }
                                    switch (elem.asType().getKind()) {
                                        case BOOLEAN:
                                            paramNames[0] = "boolean";
                                            break;
                                        case BYTE:
                                            paramNames[0] = "byte";
                                            break;
                                        case SHORT:
                                            paramNames[0] = "short";
                                            break;
                                        case INT:
                                            paramNames[0] = "int";
                                            break;
                                        case LONG:
                                            paramNames[0] = "long";
                                            break;
                                        case FLOAT:
                                            paramNames[0] = "float";
                                            break;
                                        case DOUBLE:
                                            paramNames[0] = "double";
                                            break;
                                        case CHAR:
                                            paramNames[0] = "char";
                                            break;
                                        case ARRAY:
                                            paramNames[0] = "byte[]";
                                            break;
                                        case DECLARED:
                                            paramNames[0] = elem.asType().toString();
                                            break;
                                        default:
                                            throw new IllegalArgumentException("Unsupported type: " + elem.asType().getKind());
                                    }
                                    break;
                                }
                            }
                            try (MethodCreator mc = creator.getMethodCreator(executableElement.getSimpleName().toString(),
                                    executableElement.getReturnType().toString(), paramNames)) {
                                mc.setModifiers(Modifier.ABSTRACT | Modifier.PUBLIC);
                                mc.addAnnotation(VerbClient.class);

                            }
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

}
