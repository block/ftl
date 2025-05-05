package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

import org.jboss.jandex.MethodInfo;
import org.jboss.logging.Logger;
import org.objectweb.asm.ClassReader;
import org.objectweb.asm.Opcodes;
import org.objectweb.asm.tree.AbstractInsnNode;
import org.objectweb.asm.tree.ClassNode;
import org.objectweb.asm.tree.LineNumberNode;
import org.objectweb.asm.tree.MethodNode;

import io.quarkus.deployment.dev.RuntimeUpdatesProcessor;
import io.quarkus.gizmo.DescriptorUtils;
import xyz.block.ftl.schema.v1.Position;

public class PositionUtils {

    private static final Logger LOG = Logger.getLogger(PositionUtils.class);

    public static xyz.block.ftl.language.v1.Position toError(Position position) {
        var fileName = position.getFilename();

        if (RuntimeUpdatesProcessor.INSTANCE != null) {
            for (var updates : RuntimeUpdatesProcessor.INSTANCE.getSourcesDir()) {
                Path resolved = updates.resolve(fileName);
                if (Files.exists(resolved)) {
                    fileName = resolved.toAbsolutePath().toString();
                    break;
                }
            }
        }
        return xyz.block.ftl.language.v1.Position.newBuilder()
                .setFilename(fileName)
                .setLine(position.getLine()).build();
    }

    public static Position forMethod(String projectRoot, MethodInfo method) {
        return getLineNumber(projectRoot, method.declaringClass().name().toString(), method);
    }

    public static Position forClass(String projectRoot, String className) {
        return getLineNumber(projectRoot, className, null);
    }

    static Position getLineNumber(String projectRoot, String className, MethodInfo method) {
        Position.Builder builder = Position.newBuilder();
        try {

            var cl = Thread.currentThread().getContextClassLoader();
            var cls = cl.getResource(className.replace('.', '/') + ".class");
            if (cls == null) {
                return builder.build();
            }

            try (var in = cls.openStream()) {
                ClassReader reader = new ClassReader(in);
                ClassNode clNode = new ClassNode(Opcodes.ASM9);
                reader.accept(clNode, Opcodes.ASM9);
                if (clNode.sourceFile == null) {
                    return builder.build();
                }

                int idx = className.lastIndexOf(".");
                String packagePart = "";
                if (idx != -1) {
                    packagePart = className.substring(0, idx).replaceAll("\\.", "/") + "/";
                }

                // Construct the relative path first
                String relativePath = packagePart + clNode.sourceFile;
                String finalPath = relativePath; // Default to relative path

                // Try to resolve the absolute path using Quarkus Dev Mode source dirs
                if (RuntimeUpdatesProcessor.INSTANCE != null) {
                    for (var updates : RuntimeUpdatesProcessor.INSTANCE.getSourcesDir()) {
                        Path resolved = updates.resolve(relativePath);
                        if (Files.exists(resolved)) {
                            finalPath = resolved.toAbsolutePath().toString();
                            break; // Found the absolute path
                        }
                    }
                }

                if (projectRoot != null && !projectRoot.isEmpty()) {
                    Path projectRootPath = Paths.get(projectRoot);
                    Path potentialPath = projectRootPath.resolve(finalPath);
                    Path absoluteSourcePath = Files.exists(potentialPath) ? potentialPath.toAbsolutePath() : null;
                    if (absoluteSourcePath != null) {
                        try {
                            finalPath = projectRootPath.toAbsolutePath().relativize(absoluteSourcePath).toString();
                        } catch (IllegalArgumentException e) {
                            LOG.debugf(
                                    "Could not relativize path %s against project root %s. Using absolute path. Error: %s",
                                    absoluteSourcePath, projectRootPath.toAbsolutePath(), e.getMessage());
                            finalPath = absoluteSourcePath.toString();
                        }
                    } else {
                        LOG.debugf("Could not resolve source file %s relative to project root %s or dev sources.",
                                relativePath, projectRoot);
                    }
                }

                builder.setFilename(finalPath);

                if (method != null) {
                    var descriptor = DescriptorUtils.methodSignatureToDescriptor(method.returnType().descriptor(),
                            method.parameters().stream().map(p -> p.type().descriptor()).toArray(String[]::new));
                    for (MethodNode mNode : clNode.methods) {
                        if (mNode.name.equals(method.name()) && mNode.desc.equals(descriptor)) {
                            for (AbstractInsnNode inNode : mNode.instructions) {
                                if (inNode instanceof LineNumberNode) {
                                    // We know the method won't be on the very first line, we subtract 2 to 'guess'
                                    // the correct line
                                    // This should line up with either the declaration annotation or the method
                                    // itself
                                    builder.setLine(Math.max(1, ((LineNumberNode) inNode).line - 2));
                                    return builder.build();
                                }
                            }
                        }
                    }
                }
                return builder.build();
            } catch (IOException e) {
                LOG.errorf(e, "Failed to read class %s", className);
                return builder.build();
            }
        } catch (IllegalStateException e) {
            // Ignore
            return builder.build();
        }
    }
}
