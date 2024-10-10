// Copyright 2024 Syntio Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package janitorsr

import "time"

type VersionDetails struct {
	VersionID          string    `json:"version_id,omitempty"`
	Version            string    `json:"version"`
	SchemaID           string    `json:"schema_id"`
	Specification      string    `json:"specification"`
	Description        string    `json:"description"`
	SchemaHash         string    `json:"schema_hash"`
	CreatedAt          time.Time `json:"created_at"`
	VersionDeactivated bool      `json:"version_deactivated"`
}

type registrationRequest struct {
	Description       string `json:"description"`
	Specification     string `json:"specification"`
	Name              string `json:"name"`
	SchemaType        string `json:"schema_type"`
	CompatibilityMode string `json:"compatibility_mode"`
	ValidityMode      string `json:"validity_mode"`
	GroupId           string `json:"publisher_id"`
}

type insertInfo struct {
	Id      string `json:"identification"`
	Version string `json:"version"`
	Message string `json:"message"`
}
