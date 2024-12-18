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

package net.syntio.validity.controller;

import io.apicurio.registry.rules.RuleViolation;
import io.apicurio.registry.rules.RuleViolationException;
import net.syntio.validity.Message;
import net.syntio.validity.checker.Checker;
import net.syntio.validity.dto.CheckRequestDto;
import net.syntio.validity.dto.CheckResponseDto;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;

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
        } catch (RuleViolationException e) {
            CheckResponseDto res = new CheckResponseDto(false);
            StringBuilder reasons = new StringBuilder();
            for (RuleViolation cause:e.getCauses()) {
                reasons.append(cause.getDescription()).append(" at ").append(cause.getContext()).append(";");
            }
            res.setInfo(reasons.toString());
            return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(res);
        } catch (Exception e) {
            CheckResponseDto res = new CheckResponseDto(false);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(res);
        }
    }

    @GetMapping(value = "/health")
    public ResponseEntity healthCheck() {
        return ResponseEntity.ok().build();
    }

}
