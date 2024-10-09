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
