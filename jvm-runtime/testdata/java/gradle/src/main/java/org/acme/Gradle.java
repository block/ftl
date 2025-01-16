package org.acme;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Gradle {

    @Export
    @Verb
    public String hello(String request) {
        return "Hello, " + request + "!";
    }
}
