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
import io.apicurio.registry.rules.compatibility.ProtobufCompatibilityChecker;

import java.util.List;

public class ProtobufChecker implements CompatibilityChecker {
    @Override
    public boolean testCompatibility(CompatibilityLevel level, List<ContentHandle> history, ContentHandle currentSchema) {
        ProtobufCompatibilityChecker cc =  new ProtobufCompatibilityChecker();
        return cc.testCompatibility(level, history, currentSchema).isCompatible();
    }
}
