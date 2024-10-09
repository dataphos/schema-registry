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

package net.syntio.csvvalidator.validator;
import uk.gov.nationalarchives.csv.validator.api.java.FailMessage;
import java.io.Reader;
import java.io.StringReader;
import java.util.ArrayList;
import java.util.List;

public class CsvValidator {
    public static boolean validate(String data, String schema) {
        Reader dataReader = new StringReader(data);
        Reader schemaReader = new StringReader(schema);

        List<FailMessage> messages = uk.gov.nationalarchives.csv.validator.api.java.CsvValidator.validate(dataReader, schemaReader,
                false,
                new ArrayList<>(),
                true,
                false);

        return messages.isEmpty();
    }
}
