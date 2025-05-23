package xyz.block.ftl.java.test.postgres;

public class TransactionRequest {
    private String[] items;

    public String[] getItems() {
        return items;
    }

    public TransactionRequest setItems(String[] items) {
        this.items = items;
        return this;
    }
}
