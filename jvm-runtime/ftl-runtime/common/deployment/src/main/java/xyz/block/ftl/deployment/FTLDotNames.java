package xyz.block.ftl.deployment;

import org.jboss.jandex.DotName;

import xyz.block.ftl.*;
import xyz.block.ftl.Enum;

public class FTLDotNames {

    private FTLDotNames() {

    }

    public static final DotName SECRET = DotName.createSimple(Secret.class);
    public static final DotName CONFIG = DotName.createSimple(Config.class);
    public static final DotName EXPORT = DotName.createSimple(Export.class);
    public static final DotName DATA = DotName.createSimple(Data.class);
    public static final DotName ENUM = DotName.createSimple(Enum.class);
    public static final DotName ENUM_HOLDER = DotName.createSimple(EnumHolder.class);
    public static final DotName VERB = DotName.createSimple(Verb.class);
    public static final DotName CRON = DotName.createSimple(Cron.class);
    public static final DotName TYPE_ALIAS_MAPPER = DotName.createSimple(TypeAliasMapper.class);
    public static final DotName TYPE_ALIAS = DotName.createSimple(TypeAlias.class);
    public static final DotName SUBSCRIPTION = DotName.createSimple(Subscription.class);
    public static final DotName LEASE_CLIENT = DotName.createSimple(LeaseClient.class);
    public static final DotName GENERATED_REF = DotName.createSimple(GeneratedRef.class);
    public static final DotName TOPIC = DotName.createSimple(Topic.class);
    public static final DotName FIXTURE = DotName.createSimple(Fixture.class.getName());
    public static final DotName SINGLE_PARTITION_MAPPER = DotName.createSimple(SinglePartitionMapper.class);
}
