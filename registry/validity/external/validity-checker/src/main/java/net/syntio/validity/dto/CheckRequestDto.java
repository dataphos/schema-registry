package net.syntio.validity.dto;

import net.syntio.validity.Message;

public class CheckRequestDto {
    private final Message message;

    public CheckRequestDto(String schema, String format, String mode) {
        this.message = new Message(format, schema, mode);
    }

    public Message getMessage() {
        return this.message;
    }

}
