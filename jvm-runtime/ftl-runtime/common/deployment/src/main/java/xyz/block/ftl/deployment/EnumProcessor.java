package xyz.block.ftl.deployment;

import static org.jboss.jandex.PrimitiveType.Primitive.BYTE;
import static org.jboss.jandex.PrimitiveType.Primitive.INT;
import static org.jboss.jandex.PrimitiveType.Primitive.LONG;
import static org.jboss.jandex.PrimitiveType.Primitive.SHORT;
import static xyz.block.ftl.deployment.FTLDotNames.ENUM_HOLDER;
import static xyz.block.ftl.deployment.FTLDotNames.GENERATED_REF;

import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Set;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.ClassType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.FieldInfo;
import org.jboss.jandex.PrimitiveType;
import org.jboss.jandex.Type;
import org.jboss.logging.Logger;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import xyz.block.ftl.VariantName;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.schema.v1.Decl;
import xyz.block.ftl.schema.v1.Enum;
import xyz.block.ftl.schema.v1.EnumVariant;
import xyz.block.ftl.schema.v1.Int;
import xyz.block.ftl.schema.v1.IntValue;
import xyz.block.ftl.schema.v1.StringValue;
import xyz.block.ftl.schema.v1.TypeValue;
import xyz.block.ftl.schema.v1.Value;
import xyz.block.ftl.schema.v1.Visibility;

public class EnumProcessor {

    private static final Logger log = Logger.getLogger(EnumProcessor.class);
    public static final Set<PrimitiveType.Primitive> INT_TYPES = Set.of(INT, LONG, BYTE, SHORT);

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    SchemaContributorBuildItem handleEnums(
            CombinedIndexBuildItem index,
            ProjectRootBuildItem projectRootBuildItem,
            FTLRecorder recorder,
            CommentsBuildItem commentsBuildItem) {
        var enumAnnotations = index.getIndex().getAnnotations(FTLDotNames.ENUM);
        log.debugf("Processing %d enum annotations into decls", enumAnnotations.size());
        return new SchemaContributorBuildItem(moduleBuilder -> {
            try {
                var decls = extractEnumDecls(index, projectRootBuildItem, enumAnnotations, recorder, moduleBuilder,
                        commentsBuildItem);
                for (var decl : decls) {
                    moduleBuilder.addDecls(decl);
                }
            } catch (ClassNotFoundException | NoSuchFieldException | IllegalAccessException e) {
                throw new RuntimeException(e);

            }
        });
    }

    /**
     * Extract all enums for this module, returning a Decl for each. Also registers the enums with the recorder, which
     * sets up Jackson serialization in the runtime.
     * ModuleBuilder.buildType is used, and has the side effect of adding child Decls to the module.
     */
    private List<Decl> extractEnumDecls(CombinedIndexBuildItem index, ProjectRootBuildItem projectRootBuildItem,
            Collection<AnnotationInstance> enumAnnotations, FTLRecorder recorder, ModuleBuilder moduleBuilder,
            CommentsBuildItem commentsBuildItem)
            throws ClassNotFoundException, NoSuchFieldException, IllegalAccessException {
        List<Decl> decls = new ArrayList<>();
        for (var enumAnnotation : enumAnnotations) {
            ClassInfo classInfo = enumAnnotation.target().asClass();
            Class<?> clazz = Class.forName(classInfo.name().toString(), false,
                    Thread.currentThread().getContextClassLoader());
            var isLocalToModule = !classInfo.hasDeclaredAnnotation(GENERATED_REF);

            Visibility visibility = VisibilityUtil.getVisibility(classInfo);
            if (classInfo.isEnum()) {
                // Value enum
                recorder.registerEnum(clazz);
                if (isLocalToModule) {
                    decls.add(extractValueEnum(projectRootBuildItem.getProjectRoot(), classInfo, clazz, visibility,
                            commentsBuildItem));
                }
            } else {
                var typeEnum = extractTypeEnum(index, projectRootBuildItem, moduleBuilder, classInfo, visibility,
                        commentsBuildItem);
                recorder.registerEnum(clazz, typeEnum.variantClasses);
                if (isLocalToModule) {
                    decls.add(typeEnum.decl);
                }
            }
        }
        return decls;
    }

    /**
     * Value enums are Java language enums with a single field 'value'
     */
    private Decl extractValueEnum(String projectRoot, ClassInfo classInfo, Class<?> clazz, Visibility visibility,
            CommentsBuildItem commentsBuildItem)
            throws NoSuchFieldException, IllegalAccessException {
        String name = classInfo.simpleName();
        Enum.Builder enumBuilder = Enum.newBuilder()
                .setName(name)
                .setPos(PositionUtils.forClass(projectRoot, classInfo.name().toString()))
                .setVisibility(visibility)
                .addAllComments(commentsBuildItem.getComments(name));
        FieldInfo valueField = classInfo.field("value");
        if (valueField == null) {
            throw new RuntimeException("Enum must have a 'value' field: " + classInfo.name());
        }
        Type type = valueField.type();
        xyz.block.ftl.schema.v1.Type.Builder typeBuilder = xyz.block.ftl.schema.v1.Type.newBuilder();
        if (isInt(type)) {
            typeBuilder.setInt(Int.newBuilder().build()).build();
        } else if (type.name().equals(DotName.STRING_NAME)) {
            typeBuilder.setString(xyz.block.ftl.schema.v1.String.newBuilder().build());
        } else {
            throw new RuntimeException(
                    "Enum value type must be String, int, long, short, or byte: " + classInfo.name());
        }
        enumBuilder.setType(typeBuilder.build());

        for (var constant : clazz.getEnumConstants()) {
            Field value = constant.getClass().getDeclaredField("value");
            value.setAccessible(true);
            Value.Builder valueBuilder = Value.newBuilder();
            if (isInt(type)) {
                long aLong = value.getLong(constant);
                valueBuilder.setIntValue(IntValue.newBuilder().setValue(aLong).build());
            } else {
                String aString = (String) value.get(constant);
                valueBuilder.setStringValue(StringValue.newBuilder().setValue(aString).build());
            }
            EnumVariant variant = EnumVariant.newBuilder()
                    .setName(constant.toString())
                    .setValue(valueBuilder)
                    .build();
            enumBuilder.addVariants(variant);
        }
        return Decl.newBuilder().setEnum(enumBuilder).build();
    }

    private record TypeEnum(Decl decl, List<Class<?>> variantClasses) {
    }

    /**
     * Type Enums are an interface with 1+ implementing classes. The classes may be: </br>
     * - a wrapper for a FTL native type e.g. string, [string]. Has @EnumHolder annotation </br>
     * - a class with arbitrary fields </br>
     */
    private TypeEnum extractTypeEnum(CombinedIndexBuildItem index, ProjectRootBuildItem projectRootBuildItem,
            ModuleBuilder moduleBuilder,
            ClassInfo classInfo, Visibility visibility, CommentsBuildItem commentsBuildItem) throws ClassNotFoundException {
        String projectRoot = projectRootBuildItem.getProjectRoot();
        String name = classInfo.simpleName();
        Enum.Builder enumBuilder = Enum.newBuilder()
                .setName(name)
                .setPos(PositionUtils.forClass(projectRoot, classInfo.name().toString()))
                .setVisibility(visibility)
                .addAllComments(commentsBuildItem.getComments(name));
        var variants = index.getComputingIndex().getAllKnownImplementors(classInfo.name());
        if (variants.isEmpty()) {
            throw new RuntimeException("No variants found for enum: " + enumBuilder.getName());
        }
        var variantClasses = new ArrayList<Class<?>>();
        for (var variant : variants) {
            Type variantType;
            String variantName = ModuleBuilder.classToName(variant);
            if (variant.hasAnnotation(ENUM_HOLDER)) {
                // Enum value holder class
                FieldInfo valueField = variant.field("value");
                if (valueField == null) {
                    throw new RuntimeException("Enum variant must have a 'value' field: " + variant.name());
                }
                variantType = valueField.type();
                // TODO add to variantClasses; write serialization code for holder classes
            } else {
                // Class is the enum variant type
                variantType = ClassType.builder(variant.name()).build();
                Class<?> variantClazz = Class.forName(variantType.name().toString(), false,
                        Thread.currentThread().getContextClassLoader());
                variantClasses.add(variantClazz);
            }
            if (variant.hasDeclaredAnnotation(VariantName.class)) {
                AnnotationInstance variantNameAnnotation = variant.annotation(VariantName.class);
                variantName = variantNameAnnotation.value().asString();
            }
            xyz.block.ftl.schema.v1.Type declType = moduleBuilder.buildType(variantType, visibility,
                    Nullability.NOT_NULL);
            TypeValue typeValue = TypeValue.newBuilder().setValue(declType).build();

            EnumVariant.Builder variantBuilder = EnumVariant.newBuilder()
                    .setName(variantName)
                    .setValue(Value.newBuilder().setTypeValue(typeValue).build());
            enumBuilder.addVariants(variantBuilder.build());
        }
        return new TypeEnum(Decl.newBuilder().setEnum(enumBuilder).build(), variantClasses);
    }

    private boolean isInt(Type type) {
        return type.kind() == Type.Kind.PRIMITIVE && INT_TYPES.contains(type.asPrimitiveType().primitive());
    }

}
