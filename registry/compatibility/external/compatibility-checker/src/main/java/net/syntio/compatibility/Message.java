package net.syntio.compatibility;

public class Message {
    private final String id;
    private final String format;
    private final String schema;

    public Message(String id, String format, String schema) {
        this.id = id;
        this.format = format;
        this.schema = schema;
    }


    public String getSchema() {
        return schema;
    }

    public String getID() {
        return id;
    }

    public String getFormat() {
        return format;
    }
}
