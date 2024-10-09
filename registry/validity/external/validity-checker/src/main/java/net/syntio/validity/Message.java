package net.syntio.validity;

public class Message {
    private final String schemaType;
    private final String schema;
    private final String validityLevel;

    public Message(String schemaType, String schema, String validityLevel) {
        this.schemaType = schemaType;
        this.schema = schema;
        this.validityLevel = validityLevel;
    }

    public String getSchemaType() {
        return schemaType;
    }

    public String getSchema() {
        return schema;
    }

    public String getValidityLevel() {
        return validityLevel;
    }
}
