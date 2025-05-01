package xyz.block.ftl.deployment;

import org.jboss.jandex.Type;

import io.quarkus.builder.item.MultiBuildItem;
import xyz.block.ftl.schema.v1.Visibility;

public final class TypeAliasBuildItem extends MultiBuildItem {

    final String name;
    final String module;
    final Type localType;
    final Type serializedType;
    final Visibility visibility;

    public TypeAliasBuildItem(String name, String module, Type localType, Type serializedType, Visibility exported) {
        this.name = name;
        this.module = module;
        this.localType = localType;
        this.serializedType = serializedType;
        this.visibility = exported;
    }

    public String getName() {
        return name;
    }

    public String getModule() {
        return module;
    }

    public Type getLocalType() {
        return localType;
    }

    public Type getSerializedType() {
        return serializedType;
    }

    public Visibility getVisibility() {
        return visibility;
    }
}
