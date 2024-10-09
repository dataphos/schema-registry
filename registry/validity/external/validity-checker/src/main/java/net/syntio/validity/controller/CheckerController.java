package net.syntio.validity.controller;

import net.syntio.validity.Message;
import net.syntio.validity.checker.Checker;
import net.syntio.validity.dto.CheckRequestDto;
import net.syntio.validity.dto.CheckResponseDto;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class CheckerController {
    @PostMapping(value = "/")
    public ResponseEntity<CheckResponseDto> check(@RequestBody CheckRequestDto req) {
        Message payload = req.getMessage();

        try {
            String schemaType = payload.getSchemaType();
            String schema = payload.getSchema();
            String mode = payload.getValidityLevel();

            boolean result = Checker.checkValidity(schemaType, schema, mode);
            CheckResponseDto res = new CheckResponseDto(result);
            if (result) {
                res.setInfo("Schema is valid");
                return ResponseEntity.ok(res);
            }
            res.setInfo("Schema is invalid");
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
