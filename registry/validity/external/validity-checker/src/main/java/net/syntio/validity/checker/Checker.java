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

package net.syntio.validity.checker;

import io.apicurio.registry.content.ContentHandle;
import io.apicurio.registry.rules.validity.ContentValidator;
import io.apicurio.registry.rules.validity.ValidityLevel;

import net.syntio.validity.ValidatorFactory;

import java.util.Collections;

public class Checker {

    public static boolean checkValidity(String schemaType, String schema, String mode) {
        ValidityLevel valLevel = switch (mode.toLowerCase()) {
            case "syntax-only" -> ValidityLevel.SYNTAX_ONLY;
            case "full" -> ValidityLevel.FULL;
            default -> ValidityLevel.NONE;
        };
        ContentValidator validator = ValidatorFactory.createValidator(schemaType);
        if (validator == null) { // in case ValidatorFactory returns null
          return false;
        }
        ContentHandle contentHandle = ContentHandle.create(schema);
        validator.validate(valLevel, contentHandle, Collections.emptyMap());
        return true;
    }
}
