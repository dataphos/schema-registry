# Dataphos Schema Registry - Registry component

Repository of the Dataphos Schema Registry API.


## Registry

The Registry, which itself is a database with a REST API on top, is deployed as a deployment on a Kubernetes cluster
which supports the following:
- Schema registration
- Schema updating (adding a new version of an existing schema)
- Retrieval of existing schemas (specified version or latest version)
- Deleting the whole schema or just specified versions of a schema
- Checking for schema validity (syntactically and semantically)
- Checking for schema compatibility (backward, forward, transitive)
- Schema search


The main component of the Schema Registry product is entirely independent of the implementation of the data-streaming
platform. It is implemented as a REST API that provides handles (via URL) for clients and communicates via HTTP
requests.

The worker component communicates with the REST API by sending the HTTP GET request that retrieves a message schema from
the Registry by using the necessary parameters. The message schemas themselves can be stored in any type of database (
Schema History), whether in tables like in standard SQL databases, such as Oracle or PostgreSQL, or NoSQL databases like 
MongoDB. The component itself has an interface with the database connector that can be easily modified to
work with databases that fit the client’s needs.


## Getting Started
### Prerequisites
Schema Registry components run in a Kubernetes environment. This quickstart guide will assume that you have
the ```kubectl``` tool installed and a running Kubernetes cluster on one of the major cloud providers (GCP, Azure) and a
connection with the cluster.

#### Namespace
Before deploying the Schema Registry, the namespace where the components will be deployed should be created if it
doesn't exist.

---
Open a command line tool of your choice and connect to your cluster. Create the namespace where Schema Registry will be
deployed. We will use namespace "dataphos" in this quickstart guide.

```yaml
kubectl create namespace dataphos
```

### Quick Start

Deploy Schema Registry - registry component using the following script. The required arguments are:

- the namespace
- Schema History Postgres password

#### Deployment

The script is located in the ```./scripts/registry/``` folder. from the content root. To run the script, run the
following command:

```bash
# "dataphos" is an example of the namespace name
# "p4sSw0rD" is example of the Schema History Postgres password
./sr_registry.sh dataphos p4sSw0rD
```

## Usage
Even thought the Schema Registry provides REST API for registering, updating, fetching a schema, fetching all the
versions, fetching the latest, deleting a schema, etc. We will showcase here only the requests to register, update and
fetch a schema.

### Register a schema

After the Schema Registry is deployed you will have access to its API endpoint. To register a schema, you have to send a
POST request to the endpoint ```http://schema-registry-svc:8080/schemas``` in whose body you need to provide the name of the
schema, description, schema_type, specification (the schema), compatibility and validity mode.

```
{
    "description": "new json schema for testing", 
    "schema_type": "json", 
    "specification":  "{\r\n  \"$id\": \"https://example.com/person.schema.json\",\r\n  \"$schema\": \"https://json-schema.org/draft/2020-12/schema\",\r\n  \"title\": \"Person\",\r\n  \"type\": \"object\",\r\n  \"properties\": {\r\n    \"firstName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's first name.\"\r\n    },\r\n    \"lastName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's last name.\"\r\n    },\r\n    \"age\": {\r\n      \"description\": \"Age in years which must be equal to or greater than zero.\",\r\n      \"type\": \"integer\",\r\n      \"minimum\": 0\r\n    }\r\n  }\r\n}\r\n",
    "name": "schema json",
    "compatibility_mode": "none",
    "validity_mode": "none"
}
```

or using curl:

``` 
curl -XPOST -H "Content-type: application/json" -d '{
    "description": "new json schema for testing", 
    "schema_type": "json", 
    "specification":  "{\r\n  \"$id\": \"https://example.com/person.schema.json\",\r\n  \"$schema\": \"https://json-schema.org/draft/2020-12/schema\",\r\n  \"title\": \"Person\",\r\n  \"type\": \"object\",\r\n  \"properties\": {\r\n    \"firstName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's first name.\"\r\n    },\r\n    \"lastName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's last name.\"\r\n    },\r\n    \"age\": {\r\n      \"description\": \"Age in years which must be equal to or greater than zero.\",\r\n      \"type\": \"integer\",\r\n      \"minimum\": 0\r\n    }\r\n  }\r\n}\r\n",
    "name": "schema json",
    "compatibility_mode": "none",
    "validity_mode": "none"
}' 'http://schema-registry-svc:8080/schemas/'
```

The response to the schema registration request will be:

- STATUS 201 Created
    ```
    {
        "identification": "32",
        "version": "1",
        "message": "schema successfully created"
    }
    ```

- STATUS 409 Conflict -> indicating that the schema already exists
    ```
    {
        "identification": "32",
        "version": "1",
        "message": "schema already exists at id=32"
    }
    ```

- STATUS 500 Internal Server Error -> indicating a server error, which means that either the request is not correct (
missing fields) or that the server is down.
    ```
    {
        "message": "Internal Server Error"
    }
    ``` 

### Update a schema

After the Schema Registry is registered you can update it by registering a new version under that schema ID. To update a
schema, you have to send a PUT request to the endpoint ```http://schema-registry-svc:8080/schemas/<schema_ID>``` in whose body
you need to provide the description (optional) of the version and the specification (the schema)

```
{
    "description": "added field for middle name",
    "specification": "{\r\n  \"$id\": \"https://example.com/person.schema.json\",\r\n  \"$schema\": \"https://json-schema.org/draft/2020-12/schema\",\r\n  \"title\": \"Person\",\r\n  \"type\": \"object\",\r\n  \"properties\": {\r\n    \"firstName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's first name.\"\r\n    },\r\n    \"lastName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's last name.\"\r\n    },\r\n    \"lastName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's last name.\"\r\n    },\r\n    \"age\": {\r\n      \"description\": \"Age in years which must be equal to or greater than zero.\",\r\n      \"type\": \"integer\",\r\n      \"minimum\": 0\r\n    }\r\n  }\r\n}\r\n"
}
```

or using curl:

```
curl -XPUT -H "Content-type: application/json" -d '{
    "description": "added field for middle name",
    "specification": "{\r\n  \"$id\": \"https://example.com/person.schema.json\",\r\n  \"$schema\": \"https://json-schema.org/draft/2020-12/schema\",\r\n  \"title\": \"Person\",\r\n  \"type\": \"object\",\r\n  \"properties\": {\r\n    \"firstName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's first name.\"\r\n    },\r\n    \"lastName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's last name.\"\r\n    },\r\n    \"lastName\": {\r\n      \"type\": \"string\",\r\n      \"description\": \"The person's last name.\"\r\n    },\r\n    \"age\": {\r\n      \"description\": \"Age in years which must be equal to or greater than zero.\",\r\n      \"type\": \"integer\",\r\n      \"minimum\": 0\r\n    }\r\n  }\r\n}\r\n"
}' 'http://schema-registry-svc:8080/schemas/<schema-id>'
```

The response to the schema updating request will be the same as for registering except when the updating is done
successfully it will be status 200 OK and a new version will be provided.

```
{
    "identification": "32",
    "version": "2",
    "message": "schema successfully updated"
}
```

### Fetch a schema version

To get a schema version and its relevant details, a GET request needs to be made and the endpoint needs to be:

```http://schema-registry-svc:8080/schemas/<schema-id>/versions/<schema-version>```

or using curl:

``` curl -XGET -H "Content-type: application/json" 'http://schema-registry-svc:8080/schemas/<schema-id>/versions/<schema-version>' ```

The response to the schema registration request will be:

- STATUS 200 OK
    ```
    {
        "id": "32",
        "version": "1",
        "schema_id": "32",
        "specification": "ew0KICAiJHNjaGVtYSI6ICJodHRwOi8vanNvbi1zY2hlbWEub3JnL2RyYWZ0LTA3L3NjaGVtYSIsDQogICJ0eXBlIjogIm9iamVjdCIsDQogICJ0aXRsZSI6ICJUaGUgUm9vdCBTY2hlbWEiLA0KICAiZGVzY3JpcHRpb24iOiAiVGhlIHJvb3Qgc2NoZW1hIGNvbXByaXNlcyB0aGUgZW50aXJlIEpTT04gZG9jdW1lbnQuIiwNCiAgImRlZmF1bHQiOiB7fSwNCiAgImFkZGl0aW9uYWxQcm9wZXJ0aWVzIjogdHJ1ZSwNCiAgInJlcXVpcmVkIjogWw0KICAgICJwaG9uZSINCiAgXSwNCiAgInByb3BlcnRpZXMiOiB7DQogICAgInBob25lIjogew0KICAgICAgInR5cGUiOiAiaW50ZWdlciIsDQogICAgICAidGl0bGUiOiAiVGhlIFBob25lIFNjaGVtYSIsDQogICAgICAiZGVzY3JpcHRpb24iOiAiQW4gZXhwbGFuYXRpb24gYWJvdXQgdGhlIHB1cnBvc2Ugb2YgdGhpcyBpbnN0YW5jZS4iLA0KICAgICAgImRlZmF1bHQiOiAiIiwNCiAgICAgICJleGFtcGxlcyI6IFsNCiAgICAgICAgMQ0KICAgICAgXQ0KICAgIH0sDQogICAgInJvb20iOiB7DQogICAgICAidHlwZSI6ICJpbnRlZ2VyIiwNCiAgICAgICJ0aXRsZSI6ICJUaGUgUm9vbSBTY2hlbWEiLA0KICAgICAgImRlc2NyaXB0aW9uIjogIkFuIGV4cGxhbmF0aW9uIGFib3V0IHRoZSBwdXJwb3NlIG9mIHRoaXMgaW5zdGFuY2UuIiwNCiAgICAgICJkZWZhdWx0IjogIiIsDQogICAgICAiZXhhbXBsZXMiOiBbDQogICAgICAgIDEyMw0KICAgICAgXQ0KICAgIH0NCiAgfQ0KfQ==",
        "description": "new json schema for testing",
        "schema_hash": "72966008fdcec8627a0e43c5d9a247501fc4ab45687dd2929aebf8ef3eb06ccd",
        "created_at": "2023-05-09T08:38:54.5515Z",
        "autogenerated": false
    }
    ```
- STATUS 404 Not Found -> indicating that the wrong schema ID or schema version was provided
- STATUS 500 Internal Server Error -> indicating a server error, which means that either the request is not correct (
wrong endpoint) or that the server is down.


### Other requests

|                    Description                    | Method |                               URL                               |              Headers               |               Body                |
|:-------------------------------------------------:|--------|:---------------------------------------------------------------:|:----------------------------------:|:---------------------------------:|
|                Get all the schemas                | GET    |               http://schema-registry-svc/schemas                |   Content-Type: application/json   | This request does not have a body |
|  Get all the schema versions of the specified ID  | GET    |        http://schema-registry-svc/schemas/{id}/versions         |   Content-Type: application/json   | This request does not have a body |
| Get the latest schema version of the specified ID | GET    |     http://schema-registry-svc/schemas/{id}/versions/latest     |   Content-Type: application/json   | This request does not have a body |
|    Get schema specification by id and version     | GET    | http://schema-registry-svc/schemas/{id}/versions/{version}/spec | Content-Type: application/json<br> | This request does not have a body |
|          Delete the schema under the ID           | DELETE |             http://schema-registry-svc/schemas/{id}             |   Content-Type: application/json   | This request does not have a body |
|        Delete the schema by id and version        | DELETE |   http://schema-registry-svc/schemas/{id}/versions/{version}    |   Content-Type: application/json   | This request does not have a body |


### Schema search
With schema search, users can swiftly locate relevant data schemas using a GET request and URL parameters.
```http://schema-registry-svc/schemas/search``` + 1 or more Query Parameters:

|                                           Query parameters                                           | Example                                                                                                   |
|:----------------------------------------------------------------------------------------------------:|-----------------------------------------------------------------------------------------------------------|
|                                                  id                                                  | search by id 5 <br>URL: http://schema-registry-svc/schemas/search?id=5                                    |
|                                               version                                                | search by id 5 and version 2<br>URL: http://schema-registry-svc/schemas/search?id=5&version=2             |
|                                                 type                                                 | search by type JSON <br>URL: http://schema-registry-svc/schemas/search?type=json                          |
|                                                 name                                                 | search by name "json_schema_name"<br>URL: http://schema-registry-svc/schemas/search?name=json_schema_name |
| orderBy name, type, id or version (if sort value is given but orderBy isn’t the default value is id) |
|         sort asc or desc (if orderBy value is given but sort isn’t the default value is asc)         |
|                                                limit                                                 |
|                                              attributes                                              | search by attributes crs and type <br> URL: http://schema-registry-svc/schemas/search?attributes=crs,type |




