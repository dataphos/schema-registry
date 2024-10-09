# Dataphos Schema Registry - Worker component

Repository of the Dataphos Schema Registry Worker.

## Worker

The Worker is deployed as a deployment on a Kubernetes cluster and performs the following:

- Message schema retrieval (and caching) from the Registry using message metadata
- Input message validation using the retrieved schema
- Input message transmission depending on its validation result

Before the producer starts sending messages their schema needs to be registered in the database, whether it is an
entirely new schema or a new version of an existing one. Each of the messages being sent to the input topic needs to
have its metadata enriched with the schema information, which includes the ID, version and the message format.

The role of the Worker component is to filter the messages being pushed from the input topic based on the metadata
attributes and route them to their destination. It does so with the help of the Registry component.

If the schema is registered in the database, the request sent to the Registry will return the schema specification and
the message can be successfully validated and routed to a topic for valid messages. In case of validation failure, the
message will be routed to a topic for dead letter messages.

Message brokers supported with the Worker component are:

- GCP Pub/Sub
- Azure ServiceBus
- Azure Event Hubs
- Apache Kafka
- Apache Pulsar
- NATS JetSteam

Also, the Schema registry enables the use of different protocols for producers and consumers, which ultimately enables
protocol conversion. For example, using the Schema registry protocol conversion you will be able to have a producer that
publishes messages using the Kafka protocol and a consumer that consumes messages using Pub/Sub protocol.

Providing a data schema and data the validators can determine if the given data is valid for the given schema. Data
types supported are:

- JSON
- AVRO
- Protocol Buffers
- XML
- CSV

Instead of logging metrics to standard output, the Worker component has Prometheus support for monitoring and alerting.


## Getting Started
### Prerequisites

Schema Registry components run in a Kubernetes environment. This quickstart guide will assume that you have
the ```kubectl``` tool installed and a running Kubernetes cluster on one of the major cloud providers (GCP, Azure) and a
connection with the cluster. The Kubernetes cluster node/nodes should have at least 8 GB of available RAM.

Schema Registry has multiple message broker options. This quickstart guide will assume that the publishing message
broker and the consuming message broker will be either GCP Pub/Sub, Azure ServiceBus or Kafka, and that you have
created:

- (in case of GCP Pub/Sub) service account JSON key with the appropriate roles (Pub/Sub Publisher, Pub/Sub Subscriber)
    - [link to create a service account](https://cloud.google.com/iam/docs/service-accounts-create#iam-service-accounts-create-console)
    - [link to create a JSON key](https://cloud.google.com/iam/docs/keys-create-delete)
- (in case of Azure ServiceBus) ServiceBus connection string
- (in case of Kafka) Kafka broker
    - [link to create a Kafka broker on Kubernetes](https://strimzi.io/docs/operators/0.30.0/quickstart.html)
- An input topic and subscription[^1] (The input topic refers to the topic that contains the data in its original
format)
- Valid topic and subscription[^1] (The valid topic refers to the topic where the data is stored after being validated
and serialized using a specific schema)
- Dead-letter topic and subscription[^1] (The valid topic refers to the topic where messages that could not be processed
by a consumer are stored for troubleshooting and analysis purposes)
- (optional) Prometheus server for gathering the metrics and monitoring the logs
    - can be deployed using the ```./scripts/prometheus.sh <your-namespace-name>``` command from the content root

[^1]: In case of Kafka, no subscription is required.

> **_NOTE:_**  All the deployment scripts are located in the ```./scripts``` folder from the content root.

---

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

Deploy Schema Registry - worker component using the following script. The required arguments are:

- the namespace
- Schema History Postgres password

#### GCP Deployment

The required arguments are:

- the namespace
- Producer Pub/Sub valid topic ID
- Producer Pub/Sub dead-letter topic ID
- name of the message type used by this worker (json, avro, protobuf, csv, xml)
- Consumer GCP Project ID
- Consumer Pub/Sub Subscription ID (created beforehand)
- Producer GCP Project ID


The script is located in the ```./scripts/sr-worker/``` folder from the content root. To run the script, run the
following command:

```
# "dataphos" is an example of the namespace name
# "valid-topic" is example of the valid topic name
# "dead-letter-topic" is example of the dead-letter topic name
# "json" is example of the message format name (needs to be either "json", "avro", "csv", "xml", "protobuf")
# "dataphos-project" is example of the consumer GCP project ID
# "input-topic-sub" is example of the input topic subcription name
# "dataphos-project" is example of the producer GCP project ID

./sr-worker-pubsub.sh "dataphos" "valid-topic" "dead-letter-topic" "json" "dataphos-project" "input-topic-sub" "dataphos-project" 
```

#### Azure (ServiceBus) Deployment
Required arguments are:

- the namespace
- Producer ServiceBus valid topic ID
- Producer ServiceBus dead-letter topic ID
- name of the message type used by this worker (json, avro, protobuf, csv, xml)
- Consumer ServiceBus Connection String
- Consumer ServiceBus Topic
- Consumer ServiceBus Subscription
- Producer ServiceBus Connection String

The script is located in the ```./scripts/sr-worker/``` folder. from the content root. To run the script, run the
following command:

```
# "dataphos" is an example of the namespace name
# "valid-topic" is example of the valid topic name
# "dead-letter-topic" is example of the dead-letter topic name
# "json" is example of the message format name (needs to be either "json", "avro", "csv", "xml", "protobuf")
# "Endpoint=sb://foo.servicebus.windows.net/;SharedAccessKeyName=someKeyName;SharedAccessKey=someKeyValue" is example of the consumer ServiceBus connection string (https://azurelessons.com/azure-service-bus-connection-string/)
# "input-topic" is example of the input topic name
# "input-topic-sub" is example of the input topic subcription name
# "Endpoint=sb://foo.servicebus.windows.net/;SharedAccessKeyName=someKeyName;SharedAccessKey=someKeyValue" is example of the producer ServiceBus connection string (https://azurelessons.com/azure-service-bus-connection-string/)

./sr-worker-servicebus.sh "dataphos" "valid-topic" "dead-letter-topic" "json" "Endpoint=sb://foo.servicebus.windows.net/;SharedAccessKeyName=someKeyName;SharedAccessKey=someKeyValue" "input-topic" "input-topic-sub" "Endpoint=sb://foo.servicebus.windows.net/;SharedAccessKeyName=someKeyName;SharedAccessKey=someKeyValue" 
```

#### Kafka Deployment (Platform agnostic)

Required arguments are:

- the namespace
- Producer Kafka valid topic ID
- Producer Kafka dead-letter topic ID
- name of the message type used by this worker (json, avro, protobuf, csv, xml)
- Consumer Kafka bootstrap server address
- Consumer Kafka Topic
- Consumer Kafka Group ID
- Producer Kafka bootstrap server address

The script is located in the ```./scripts/sr-worker/``` folder. from the content root. To run the script, run the
following command:

```
# "dataphos" is an example of the namespace name
# "valid-topic" is example of the valid topic name
# "dead-letter-topic" is example of the dead-letter topic name
# "json" is example of the message format name (needs to be either "json", "avro", "csv", "xml", "protobuf")
# "127.0.0.1:9092" is example of the consumer Kafka bootstrap server address
# "input-topic" is example of the input topic name
# "group01" is example of the input topic group ID
# "127.0.0.1:9092" is example of the producer Kafka bootstrap server address

./sr-worker-kafka.sh "dataphos" "valid-topic" "dead-letter-topic" "json" "127.0.0.1:9092" "input-topic" "group01" "127.0.0.1:9092" 
```

## Usage


### Message format

Depending on the technology your producer uses, the way you shape the message may differ and therefore the part of the
message that contains the metadata might be called ```attributes```, ```metadata,``` etc.

Besides the data field, which contains the message data, inside the attributes (or metadata) structure it's important to
add fields ```schemaId```, ```versionId``` and ```format```
which are important information for the worker component. In case some additional attributes are provided, the worker
won't lose them, they will be delegated to the destination topic.

### GCP Pub/Sub

```
{
    "ID": string,
    "Data": string,
    "Attributes": {
        schemaId: string,
        versionId: string,
        format: string,
        ...
    },
    "PublishTime": time,
}
```

| Field      | Description                                                                                                                                                                                                                                                                                                            |
|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Data       | **string** (bytes format)<br><br>The message data field. If this field is empty, the message must contain at least one attribute.<br><br>A base64-encoded string.                                                                                                                                                          |
| Attributes | **map** (key: string, value: string)<br><br>Attributes for this message. If this field is empty, the message must contain non-empty data. This can be used to filter messages on the subscription.<br><br>An object containing a list of "key": value pairs. Example: { "schemaId": "1", "versionId": "2", "format": "json" }. |
| PublishTime| **time** (time.Time format) <br><br>PublishTime is the time at which the message was published. This is populated by the server for Messages obtained from a subscription.|

### Azure ServiceBus
```
{
    "MessageID": string,
    "Body": string,
    "PartitionKey": string, 
    "ApplicationProperties": {
        schemaId: string,
        versionId: string,
        format: string,
        ...
    },
    EnqueuedTime: time
}
```

| Field      | Description                                                                                                                                                                                                                                                                                                            |
|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Body       | **string** (bytes format)<br><br>The message data field. If this field is empty, the message must contain at least one application property.                                                                                                                                                         |
| ApplicationProperties | **map** (key: string, value: string)<br><br>Attributes for this message. ApplicationProperties can be used to store custom metadata for a message.<br><br>An object containing a list of "key": value pairs. Example: { "schemaId": "1", "versionId": "2", "format": "json" }. |
| PartitionKey| **string** <br><br>PartitionKey is used with a partitioned entity and enables assigning related messages to the same internal partition. This ensures that the submission sequence order is correctly recorded. The partition is chosen by a hash function in Service Bus and cannot be chosen directly.|
| EnqueuedTime| **time** (time.Time format) <br><br>EnqueuedTime is the UTC time when the message was accepted and stored by Service Bus.|


### Kafka

```
{
    "Key": string, 
    "Value": string, 
    "Offset": int64,
    "Partition": int32,
    "Headers": {
        schemaId: string,
        versionId: string,
        format: string,
        ...
    },
    Timestamp: time
}
```

| Field      | Description                                                                                                                                                                                                                                                                                                            |
|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Key       | **string** (bytes format)<br><br>Key is an optional field that can be used for partition assignment.                                                                                                                                                         |
| Value       | **string** (bytes format)<br><br>Value is blob of data to write to Kafka.                                                                                                                                                       |
| Offset | **int64** <br><br> Offset is the offset that a record is written as.|
| Partition | **int32** <br><br> Partition is the partition that a record is written to.|
| Headers | **map** (key: string, value: string)<br><br>Headers are optional key/value pairs that are passed along with records.<br><br>Example: { "schemaId": "1", "versionId": "2", "format": "json" }. <br><br> These are purely for producers and consumers; Kafka does not look at this field and only writes it to disk. |
| Timestamp| **time** (time.Time format) <br><br>Timestamp is the timestamp that will be used for this record. Record batches are always written with "CreateTime", meaning that timestamps are generated by clients rather than brokers.|

