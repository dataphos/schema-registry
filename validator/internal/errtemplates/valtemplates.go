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

package errtemplates

import "github.com/pkg/errors"

const (
	requiredTagFailTemplate     = "Validation for '%s' failed: can not be blank"
	fileTagFailTemplate         = "Validation for '%s' failed: '%s' does not exist"
	urlTagFailTemplate          = "Validation for '%s' failed: '%s' incorrect url"
	oneofTagFailTemplate        = "Validation for '%s' failed: '%s' is not one of the options"
	hostnamePortTagFailTemplate = "Validation for '%s' failed: '%s' incorrect hostname and port"
)

func RequiredTagFail(cause string) error {
	return errors.Errorf(requiredTagFailTemplate, cause)
}

func FileTagFail(cause string, value interface{}) error {
	return errors.Errorf(fileTagFailTemplate, cause, value)
}

func UrlTagFail(cause string, value interface{}) error {
	return errors.Errorf(urlTagFailTemplate, cause, value)
}

func OneofTagFail(cause string, value interface{}) error {
	return errors.Errorf(oneofTagFailTemplate, cause, value)
}

func HostnamePortTagFail(cause string, value interface{}) error {
	return errors.Errorf(hostnamePortTagFailTemplate, cause, value)
}
