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

package net.syntio.compatibility.dto;

import net.syntio.compatibility.Message;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.List;

public class CheckRequestDto {
    private Message message;
    private final List<String> history;
    private final String mode;

    public CheckRequestDto(String payload, List<String> history, String mode) {
        try {
            this.message = transformStringToMessage(payload);
        } catch (Exception e) {
            this.message = new Message("", "", "");
            System.err.println("Cannot read message");
        }
        this.history = history;
        this.mode = mode;
    }

    public Message getMessage() {
        return message;
    }

    public String getSchema() {
        return message.getSchema();
    }

    public List<String> getHistory() {
        return history;
    }

    public String getMode() {
        return mode;
    }

    private static Message transformStringToMessage(String payload) throws JSONException {
        JSONObject jsonObject = new JSONObject(payload);
        String id = jsonObject.getString("id");
        String format = jsonObject.getString("format");
        String newSchema = jsonObject.getString("schema");
        return new Message(id, format, newSchema);
    }
}
