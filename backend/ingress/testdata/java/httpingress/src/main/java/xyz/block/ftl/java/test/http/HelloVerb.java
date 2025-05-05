package xyz.block.ftl.java.test.http;

import xyz.block.ftl.SourceVerb;
import xyz.block.ftl.Verb;

@Verb
public class HelloVerb implements SourceVerb<String> {
    @Override
    public String call() {
        return "Hello";
    }
}
