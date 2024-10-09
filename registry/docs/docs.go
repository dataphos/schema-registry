// Code generated by swaggo/swag. DO NOT EDIT.

package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
	"description": "{{escape .Description}}",
	"title": "{{.Title}}",
	"contact": {},
	"version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
	"/schemas": {
	    "get": {
		"produces": [
		    "application/json"
		],
		"summary": "Get all active schemas",
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    },
	    "post": {
		"consumes": [
		    "application/json"
		],
		"produces": [
		    "application/json"
		],
		"summary": "Post new schema",
		"parameters": [
		    {
			"description": "schema registration request",
			"name": "data",
			"in": "body",
			"schema": {
			    "$ref": "#/definitions/registry.SchemaRegistrationRequest"
			}
		    }
		],
		"responses": {
		    "201": {
			"description": "Created"
		    },
		    "400": {
			"description": "Bad Request"
		    },
		    "409": {
			"description": "Conflict"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    }
	},
	"/schemas/all": {
	    "get": {
		"produces": [
		    "application/json"
		],
		"summary": "Get all schemas",
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    }
	},
	"/schemas/search": {
	    "get": {
		"produces": [
		    "application/json"
		],
		"summary": "Search schemas",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "query"
		    },
		    {
			"type": "string",
			"description": "schema version",
			"name": "version",
			"in": "query"
		    },
		    {
			"type": "string",
			"description": "schema type",
			"name": "type",
			"in": "query"
		    },
		    {
			"type": "string",
			"description": "schema name",
			"name": "name",
			"in": "query"
		    },
		    {
			"type": "string",
			"description": "order by name, type, id or version",
			"name": "orderBy",
			"in": "query"
		    },
		    {
			"type": "string",
			"description": "sort schemas either asc or desc",
			"name": "sort",
			"in": "query"
		    },
		    {
			"type": "string",
			"description": "maximum number of retrieved schemas matching the criteria",
			"name": "limit",
			"in": "query"
		    },
		    {
			"type": "string",
			"description": "schema attributes",
			"name": "attributes",
			"in": "query"
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "400": {
			"description": "Bad Request"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    }
	},
	"/schemas/{id}": {
	    "put": {
		"consumes": [
		    "application/json"
		],
		"produces": [
		    "application/json"
		],
		"summary": "Put new schema version",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "path",
			"required": true
		    },
		    {
			"description": "schema update request",
			"name": "data",
			"in": "body",
			"required": true,
			"schema": {
			    "$ref": "#/definitions/registry.SchemaUpdateRequest"
			}
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "400": {
			"description": "Bad Request"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "409": {
			"description": "Conflict"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    },
	    "delete": {
		"produces": [
		    "application/json"
		],
		"summary": "Delete schema by schema id",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "path",
			"required": true
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "400": {
			"description": "Bad Request"
		    },
		    "404": {
			"description": "Not Found"
		    }
		}
	    }
	},
	"/schemas/{id}/versions": {
	    "get": {
		"produces": [
		    "application/json"
		],
		"summary": "Get all active schema versions by schema id",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "path",
			"required": true
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    }
	},
	"/schemas/{id}/versions/all": {
	    "get": {
		"produces": [
		    "application/json"
		],
		"summary": "Get schema by schema id",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "path",
			"required": true
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    }
	},
	"/schemas/{id}/versions/latest": {
	    "get": {
		"produces": [
		    "application/json"
		],
		"summary": "Get the latest schema version by schema id",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "path",
			"required": true
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    }
	},
	"/schemas/{id}/versions/{version}": {
	    "get": {
		"produces": [
		    "application/json"
		],
		"summary": "Get schema version by schema id and version",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "path",
			"required": true
		    },
		    {
			"type": "string",
			"description": "version",
			"name": "version",
			"in": "path",
			"required": true
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    },
	    "delete": {
		"consumes": [
		    "application/json"
		],
		"produces": [
		    "application/json"
		],
		"summary": "Delete schema version by schema id and version",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "path",
			"required": true
		    },
		    {
			"type": "string",
			"description": "version",
			"name": "version",
			"in": "path",
			"required": true
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "400": {
			"description": "Bad Request"
		    },
		    "404": {
			"description": "Not Found"
		    }
		}
	    }
	},
	"/schemas/{id}/versions/{version}/spec": {
	    "get": {
		"produces": [
		    "application/json"
		],
		"summary": "Get schema specification by schema id and version",
		"parameters": [
		    {
			"type": "string",
			"description": "schema id",
			"name": "id",
			"in": "path",
			"required": true
		    },
		    {
			"type": "string",
			"description": "version",
			"name": "version",
			"in": "path",
			"required": true
		    }
		],
		"responses": {
		    "200": {
			"description": "OK"
		    },
		    "404": {
			"description": "Not Found"
		    },
		    "500": {
			"description": "Internal Server Error"
		    }
		}
	    }
	}
    },
    "definitions": {
	"registry.SchemaRegistrationRequest": {
	    "type": "object",
	    "properties": {
		"attributes": {
		    "type": "string"
		},
		"compatibility_mode": {
		    "type": "string"
		},
		"description": {
		    "type": "string"
		},
		"last_created": {
		    "type": "string"
		},
		"name": {
		    "type": "string"
		},
		"publisher_id": {
		    "type": "string"
		},
		"schema_type": {
		    "type": "string"
		},
		"specification": {
		    "type": "string"
		},
		"validity_mode": {
		    "type": "string"
		}
	    }
	},
	"registry.SchemaUpdateRequest": {
	    "type": "object",
	    "properties": {
		"attributes": {
		    "type": "string"
		},
		"description": {
		    "type": "string"
		},
		"specification": {
		    "type": "string"
		}
	    }
	}
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Schema Registry API",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
