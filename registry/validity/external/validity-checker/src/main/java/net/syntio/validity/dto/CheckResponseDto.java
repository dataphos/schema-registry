package net.syntio.validity.dto;

public class CheckResponseDto {
    private final boolean result;
    private String info;

    public CheckResponseDto(boolean result) {
        this.result = result;
    }

    public boolean getResult() {
        return result;
    }

    public String getInfo() {
        return info;
    }

    public void setInfo(String info) {
        this.info = info;
    }
}
