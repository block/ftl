package xyz.block.ftl.runtime;

import jakarta.enterprise.context.RequestScoped;

import io.quarkus.arc.Arc;
import xyz.block.ftl.v1.Metadata;
import xyz.block.ftl.v1.Metadata.Pair;

@RequestScoped
public class CurrentTransaction {

    private static final String METADATA_KEY = "ftl-transaction";

    private String id;

    public void setId(String id) {
        this.id = id;
    }

    public String getId() {
        return id;
    }

    public void clear() {
        this.id = null;
    }

    public static CurrentTransaction current() {
        return Arc.container().select(CurrentTransaction.class).get();
    }

    public static void setCurrentId(String id) {
        if (id != null) {
            CurrentTransaction ctx = current();
            if (ctx != null) {
                ctx.setId(id);
            }
        }
    }

    public static String getCurrentId() {
        CurrentTransaction ctx = current();
        return ctx != null ? ctx.getId() : null;
    }

    public static void clearCurrent() {
        CurrentTransaction ctx = current();
        if (ctx != null) {
            ctx.clear();
        }
    }

    public static void setCurrentIdFromMetadata(Metadata metadata) {
        if (metadata != null) {
            for (var pair : metadata.getValuesList()) {
                if (pair.getKey().equals(METADATA_KEY)) {
                    setCurrentId(pair.getValue());
                }
            }
        }
    }

    public static Metadata getMetadataWithCurrentId() {
        var builder = Metadata.newBuilder();
        if (getCurrentId() != null) {
            builder.addValues(Pair.newBuilder().setKey(METADATA_KEY).setValue(getCurrentId()).build());
        }
        return builder.build();
    }

}
