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
import io.apicurio.registry.rules.compatibility.CompatibilityDifference;
import io.apicurio.registry.rules.compatibility.CompatibilityExecutionResult;
import io.apicurio.registry.rules.compatibility.CompatibilityLevel;
import io.apicurio.registry.rules.compatibility.JsonSchemaCompatibilityChecker;

import java.util.ArrayList;
import java.util.List;
import java.util.Set;

public class JsonChecker implements CompatibilityChecker {

    @Override
    public List<String> testCompatibility(CompatibilityLevel level, List<ContentHandle> history, ContentHandle currentSchema) {
        JsonSchemaCompatibilityChecker cc =  new JsonSchemaCompatibilityChecker();
        CompatibilityExecutionResult res = cc.testCompatibility(level, history, currentSchema);
        Set<CompatibilityDifference> resSet = res.getIncompatibleDifferences();
        List<String> issuesList = new ArrayList<>();
        for (CompatibilityDifference dif : resSet) {
            issuesList.add(dif.asRuleViolation().getDescription());
        }
        return issuesList;
    }
}
