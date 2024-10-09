package net.syntio.csvvalidator.dto;

public class ValidatorResponseDto {
    private final boolean validation;
    private  String info;

    public ValidatorResponseDto(boolean validation) {
        this.validation = validation;
    }

    public boolean getValidation() {
        return validation;
    }

    public String getInfo() {
        return info;
    }

    public void setInfo(String info) {
        this.info = info;
    }
}
