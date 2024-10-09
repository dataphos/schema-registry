package net.syntio.csvvalidator.controller;

import net.syntio.csvvalidator.dto.ValidationRequestDto;
import net.syntio.csvvalidator.dto.ValidatorResponseDto;
import net.syntio.csvvalidator.validator.CsvValidator;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class ValidatorController {

    @PostMapping(value = "/")
    public ResponseEntity<ValidatorResponseDto> validate(@RequestBody ValidationRequestDto req) {
        String data = req.getData().replaceAll("\r\n", "\n");
        String schema = req.getSchema().replaceAll("\r\n", "\n");
        try {
            boolean validation = CsvValidator.validate(data, schema);
            ValidatorResponseDto res = new ValidatorResponseDto(validation);
            if (validation) {
                res.setInfo("Data is valid");
                return ResponseEntity.ok(res);
            }
            res.setInfo("Data is invalid");
            return ResponseEntity.ok(res);
        } catch (Exception e) {
            return ResponseEntity.badRequest().build();
        }
    }

    @GetMapping(value = "/health")
    public ResponseEntity healthCheck() {
        return ResponseEntity.ok().build();
    }
}
