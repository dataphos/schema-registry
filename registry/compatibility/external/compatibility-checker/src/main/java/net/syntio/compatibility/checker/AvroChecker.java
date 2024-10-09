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

package net.syntio.compatibility.checker;

import io.apicurio.registry.content.ContentHandle;
import io.apicurio.registry.rules.compatibility.CompatibilityLevel;
import io.confluent.kafka.schemaregistry.avro.AvroSchema;
import io.confluent.kafka.schemaregistry.CompatibilityChecker;

import java.util.ArrayList;
import java.util.List;

public class AvroChecker implements net.syntio.compatibility.checker.CompatibilityChecker {
    @Override
    public boolean testCompatibility(CompatibilityLevel level, List<ContentHandle> history, ContentHandle currentSchema) {
        io.confluent.kafka.schemaregistry.CompatibilityLevel avroCompatibilityLevel = switch (level) {
            case NONE -> io.confluent.kafka.schemaregistry.CompatibilityLevel.NONE;
            case BACKWARD -> io.confluent.kafka.schemaregistry.CompatibilityLevel.BACKWARD;
            case BACKWARD_TRANSITIVE -> io.confluent.kafka.schemaregistry.CompatibilityLevel.BACKWARD_TRANSITIVE;
            case FORWARD -> io.confluent.kafka.schemaregistry.CompatibilityLevel.FORWARD;
            case FORWARD_TRANSITIVE -> io.confluent.kafka.schemaregistry.CompatibilityLevel.FORWARD_TRANSITIVE;
            case FULL -> io.confluent.kafka.schemaregistry.CompatibilityLevel.FULL;
            case FULL_TRANSITIVE -> io.confluent.kafka.schemaregistry.CompatibilityLevel.FULL_TRANSITIVE;
        };
        List<AvroSchema> newHistory = new ArrayList<>();
        for (ContentHandle existingArtifact : history) {
            newHistory.add(new AvroSchema(existingArtifact.content()));
        }
        AvroSchema newSchema = new AvroSchema(currentSchema.content());

        List<String> issues = switch (avroCompatibilityLevel) {
            case BACKWARD -> CompatibilityChecker.BACKWARD_CHECKER.isCompatible(newSchema, newHistory);
            case BACKWARD_TRANSITIVE -> CompatibilityChecker.BACKWARD_TRANSITIVE_CHECKER.isCompatible(newSchema, newHistory);
            case FORWARD -> CompatibilityChecker.FORWARD_CHECKER.isCompatible(newSchema, newHistory);
            case FORWARD_TRANSITIVE -> CompatibilityChecker.FORWARD_TRANSITIVE_CHECKER.isCompatible(newSchema, newHistory);
            case FULL -> CompatibilityChecker.FULL_CHECKER.isCompatible(newSchema, newHistory);
            case FULL_TRANSITIVE -> CompatibilityChecker.FULL_TRANSITIVE_CHECKER.isCompatible(newSchema, newHistory);
            case NONE -> CompatibilityChecker.NO_OP_CHECKER.isCompatible(newSchema, newHistory);
        };
        return issues.isEmpty();
    }
}
