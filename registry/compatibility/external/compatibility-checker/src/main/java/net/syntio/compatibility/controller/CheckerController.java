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

package net.syntio.compatibility.controller;

import io.apicurio.registry.rules.compatibility.CompatibilityLevel;
import net.syntio.compatibility.Message;
import net.syntio.compatibility.checker.Checker;
import net.syntio.compatibility.dto.CheckRequestDto;
import net.syntio.compatibility.dto.CheckResponseDto;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;

import java.util.ArrayList;
import java.util.List;

@RestController
public class CheckerController {
    @PostMapping(value = "/")
    public ResponseEntity<CheckResponseDto> check(@RequestBody CheckRequestDto req) {
        Message latestSchema = req.getMessage();
        List<String> schemaHistory = req.getHistory();
        CheckResponseDto res;
        try {
            schemaHistory.replaceAll(s -> s.replaceAll("\r\n", "\n"));
            String mode = req.getMode();

            CompatibilityLevel cl = getCompatibilityLevel(mode);
            List<String> issues;
            if (cl.equals(CompatibilityLevel.NONE)) {
              issues = new ArrayList<>();
            } else {
              issues = Checker.checkCompatibility(latestSchema, schemaHistory, cl);
            }

            for (String issue:issues) {
                System.out.println(issue);
            }
            res = new CheckResponseDto(issues.isEmpty());
            if (issues.isEmpty()) {
                res.setInfo("Schema is compatible");
                return ResponseEntity.ok(res);
            }
            res.setInfo("Schema is incompatible: " + String.join("; ", issues));
            return ResponseEntity.ok(res);
        } catch (NullPointerException e) {
            res = new CheckResponseDto(false);
            res.setInfo("schema history is null");
            return ResponseEntity.status(HttpStatus.PRECONDITION_FAILED).body(res);
        } catch (org.everit.json.schema.SchemaException e) {
            res = new CheckResponseDto(false);
            res.setInfo("schema version unknown or unsupported");
            return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(res);
        } catch (Exception e) {
            res = new CheckResponseDto(false);
            res.setInfo("unknown error");
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(res);
        }
    }

    @GetMapping(value = "/health")
    public ResponseEntity healthCheck() {
        return ResponseEntity.ok().build();
    }

    private CompatibilityLevel getCompatibilityLevel(String mode) throws Exception {
        return switch (mode.toUpperCase()) {
            case "BACKWARD" -> CompatibilityLevel.BACKWARD;
            case "BACKWARD_TRANSITIVE" -> CompatibilityLevel.BACKWARD_TRANSITIVE;
            case "FORWARD" -> CompatibilityLevel.FORWARD;
            case "FORWARD_TRANSITIVE" -> CompatibilityLevel.FORWARD_TRANSITIVE;
            case "FULL" -> CompatibilityLevel.FULL;
            case "FULL_TRANSITIVE" -> CompatibilityLevel.FULL_TRANSITIVE;
            case "NONE", "" -> CompatibilityLevel.NONE;
            default -> throw new Exception("Unknown compatibility mode");
        };
    }

}
