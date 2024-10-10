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

package errcodes

const (
	DatabaseConnectionInitialization = 100
	InvalidDatabaseState             = 101
	DatabaseInitialization           = 102
	ServerInitialization             = 103
	ExternalCheckerInitialization    = 104
	ServerShutdown                   = 200
	BadRequest                       = 400
	InternalServerError              = 500
	Miscellaneous                    = 999
)

func FromHttpStatusCode(code int) uint64 {
	switch {
	case code >= 400 && code < 500:
		return BadRequest
	case code >= 500:
		return InternalServerError
	default:
		return Miscellaneous
	}
}
