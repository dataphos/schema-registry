package net.syntio.csvvalidator.dto;

public class ValidationRequestDto {
    private final String data;
    private final String schema;

    public ValidationRequestDto(String data, String schema) {
        this.data = data;
        this.schema = schema;
    }

    public String getData() {
        return data;
    }

    public String getSchema() {
        return schema;
    }
}
