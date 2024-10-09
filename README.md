# Schema Registry

[![Apache 2.0 License](https://img.shields.io/github/license/dataphos/schema-registry)](./LICENSE) 
[![GitHub Release](https://img.shields.io/github/v/release/dataphos/schema-registry?sort=semver)](https://github.com/dataphos/schema-registry/releases/latest)

Schema Registry is a product used for **schema management** and **message validation**.

Schema management itself consists of 2 steps -  *schema registration* and *schema versioning*, while message validation consists of validators that validate messages for the given message schema. The core components are a server with HTTP RESTful interface used to manage the schemas, and lightweight message validators, which verify the schema and validity of the incoming messages.

It allows developers to define and manage standard schemas for events, share them across the organization, evolve the schemas while preserving compatibility, as well as validate events with the given event schema. For each schema used, the product stores their own versioned history while also providing an easy-to-use RESTful interface to work with them.

Apart from the general idea of the product, its main features are split across two major components - Registry and Validator.

The official Schema Registry documentation is available [here](https://docs.dataphos.com/schema_registry/). It contains an in-depth overview of each component, a quickstart setup, detailed deployment instructions, configuration options, usage guides and more, so be sure to check it out for better understanding.

## Registry
The Registry component represents the main database, called Schema History, which is used for handling schemas, and the REST API on top of the database to enable the other major component, [Validator](#validator), to fetch all the necessary information regarding the schemas.

## Validator
The Validator component, in essence, does what the title suggests - validates messages. It performs this by retrieving and caching messages schemas from the [Registry](#registry) database, using a message's metadata.

## Installation
In order to use Schema Registry as a whole with both of its components, the only major requirement from the user is to have a running project on one of the two major cloud providers: GCP or Azure.

All of the other requirements for the product to fully-function (message broker instance, incoming message type definition, identity and access management of the particular cloud) are further explained and can be analyzed in the [Quickstart section](https://docs.dataphos.com/schema_registry/quickstart/) of its official documentation.

## Usage
### Registry
- Takes care of everything related to the schemas themselves - registration, updates, retrieval, deletion of an entire schema or its particular version, as well as performing schema checks for validity and compatibility (backwards, forwards and transitively).
- Its REST API provides handles for clients and communicates via HTTP requests.
- With regards to the message schemas themselves, the Schema History database where they get stored in can be anything from a standard SQL database like Oracle or PostgreSQL, to a NoSQL database like Firestore or MongoDB.

### Validator
- In order for the Validator to work, the message schema needs to be registered in the Schema History database.
  - Each of the incoming messages needs to have its metadata enriched with the information of the schema stored in the Schema History database, with the main attributes being the *ID*, the *schema version* and the *message format*.
- Once that scenario is set up, the Validator can then filter incoming messages and route them to the appropriate destination - valid topic for successfully validated messages and dead-letter topic for unsucessfully validated messages.
  -  The list of supported message brokers can be found in the [Validator section](https://docs.dataphos.com/schema_registry/what-is-schema-registry/#worker) of its official documentation.
- Similarly to the various message brokers, the Validator also enables the use of different protocols for producers and consumers of messages.
  - This in turn enables protocol conversion through the system.
  - The list of supported protocols can also be found in the [Validator section](https://docs.dataphos.com/schema_registry/what-is-schema-registry/#worker) of its official documentation.

## Contributing
For all the inquiries regarding contributing to the project, be sure to check out the information in the [CONTRIBUTING.md](CONTRIBUTING.md) file.

## License
This project is licensed under the [Apache 2.0 License](LICENSE).