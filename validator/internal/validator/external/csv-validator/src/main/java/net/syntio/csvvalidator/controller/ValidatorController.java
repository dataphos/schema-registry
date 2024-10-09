/*
 * Copyright 2024 Syntio Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
