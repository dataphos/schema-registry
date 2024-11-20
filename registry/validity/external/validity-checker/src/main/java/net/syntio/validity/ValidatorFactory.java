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

package net.syntio.validity;

import io.apicurio.registry.rules.validity.AvroContentValidator;
import io.apicurio.registry.rules.validity.ContentValidator;
import io.apicurio.registry.rules.validity.JsonSchemaContentValidator;
import io.apicurio.registry.rules.validity.ProtobufContentValidator;
import io.apicurio.registry.rules.validity.XsdContentValidator;

public class ValidatorFactory {

    public static ContentValidator createValidator(String schema) {
        return switch (schema) {
            case SchemaTypes.JSON -> new JsonSchemaContentValidator();
            case SchemaTypes.AVRO -> new AvroContentValidator();
            case SchemaTypes.PROTOBUF -> new ProtobufContentValidator();
            case SchemaTypes.XML -> new XsdContentValidator();
            default -> null;
        };
    }

}
