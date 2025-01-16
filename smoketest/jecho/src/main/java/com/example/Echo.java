package com.example;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Echo {

    @Export
    @Verb
    public String hello(String request) {
        return "Hello, " + request + "!";
    }
}
