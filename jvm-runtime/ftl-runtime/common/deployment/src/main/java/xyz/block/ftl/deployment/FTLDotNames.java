package xyz.block.ftl.deployment;

import org.jboss.jandex.DotName;

import xyz.block.ftl.*;
import xyz.block.ftl.Enum;

public class FTLDotNames {

    private FTLDotNames() {

    }

    public static final DotName SECRET = DotName.createSimple(Secret.class);
    public static final DotName CONFIG = DotName.createSimple(Config.class);
    public static final DotName EGRESS = DotName.createSimple(Egress.class);
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
    public static final DotName WORKLOAD_IDENTITY = DotName.createSimple(WorkloadIdentity.class);
    public static final DotName GENERATED_REF = DotName.createSimple(GeneratedRef.class);
    public static final DotName TOPIC = DotName.createSimple(Topic.class);
    public static final DotName FIXTURE = DotName.createSimple(Fixture.class.getName());
    public static final DotName SINGLE_PARTITION_MAPPER = DotName.createSimple(SinglePartitionMapper.class);
    public static final DotName TRANSACTIONAL = DotName.createSimple(Transactional.class);
    public static final DotName KOTLIN_UNIT = DotName.createSimple("kotlin.Unit");
    public static final DotName KOTLIN_UBYTE = DotName.createSimple("kotlin.UByte");
    public static final DotName KOTLIN_USHORT = DotName.createSimple("kotlin.UShort");
    public static final DotName KOTLIN_UINT = DotName.createSimple("kotlin.UInt");
    public static final DotName KOTLIN_ULONG = DotName.createSimple("kotlin.ULong");
    public static final DotName EMPTY_VERB = DotName.createSimple(EmptyVerb.class);
    public static final DotName SOURCE_VERB = DotName.createSimple(SourceVerb.class);
    public static final DotName SINK_VERB = DotName.createSimple(SinkVerb.class);
    public static final DotName FUNCTION_VERB = DotName.createSimple(FunctionVerb.class);
}
