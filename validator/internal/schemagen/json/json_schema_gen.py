# Copyright 2024 Syntio Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json

import sys
from genson import SchemaBuilder


def main():
    data_file = open(sys.argv[1]) if len(sys.argv) > 1 else sys.stdin
    data = data_file.read()
    to_add = {"additionalProperties": False}
    builder = SchemaBuilder()
    try:
        builder.add_schema(to_add)
        builder.add_object(json.loads(data))
        result_schema = builder.to_json(indent=2)
    except ValueError:
        result_schema = ""
    sys.stdout.write(result_schema)


if __name__ == "__main__":
    main()
