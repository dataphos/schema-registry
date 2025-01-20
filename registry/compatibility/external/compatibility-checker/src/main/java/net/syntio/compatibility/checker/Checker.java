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
import net.syntio.compatibility.Message;
import net.syntio.compatibility.CheckerFactory;

import java.util.ArrayList;
import java.util.List;

public class Checker {
    public static List<String> checkCompatibility(Message msg, List<String> history, CompatibilityLevel mode) throws Exception {
        ContentHandle schema = ContentHandle.create(msg.getSchema());
        List<ContentHandle> schemaHistory = new ArrayList<>();
        for (String s : history) {
            ContentHandle ps = ContentHandle.create(s);
            schemaHistory.add(ps);
        }
        CompatibilityChecker cc = CheckerFactory.createChecker(msg.getFormat().toLowerCase());
        return cc.testCompatibility(mode, schemaHistory, schema);
    }
}
