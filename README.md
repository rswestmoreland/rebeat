# Rebeat

## Description

This application is used as a Relay/Repeater (hence the name Rebeat) for  messages sent by other Elastic Beats via Lumberjack v2 (Beats Protocol) which is used for Logstash destinations.

Ensure that this folder is at the following location:
`${GOPATH}/github.com/rswestmoreland`

## Getting Started with Rebeat

### Configuration Options

- `rebeat.address` : The address on which the process will listen (Default: 127.0.0.1)
- `rebeat.port` : The port on which the process will listen (Default = 5044)
- `rebeat.max_message_size` : The maximum accepted message size (Default = 8192)
- `rebeat.json_mode`: Enable logging of only JSON formatted messages (Default = true)
- `rebeat.default_es_log_type`: Elasticsearch type to assign to an event if one isn't specified (Default: rebeat)
- `rebeat.enable_json_validation` : Boolean value indicating if JSON schema validation should be applied for `json` format messages (Default = false)
- `rebeat.validate_all_json_types` : When json_mode enabled, indicates if ALL types must have a schema specified. Log entries with types that have no schema will not be published. (Default = false)
- `rebeat.json_schema` :  A hash consisting of the Elasticsearch type as the key, and the absolute local schema file path as the value.

### Configuration Example

The following are examples of configuration blocks for the `rebeat` section.  

1. [Configuration](_sample/config1.yml) TBD
2. [Configuration](_sample/config2.yml) TBD


#### Considerations

TBD


### Requirements

* [Golang](https://golang.org/dl/) 1.7

### Init Project
To get running with Rebeat and also install the
dependencies, run the following command:

```
make setup
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Rebeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/rswestmoreland/rebeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Rebeat run the command below. This will generate a binary
in the same directory with the name rebeat.

```
make
```

If you'd like to build the binary for OSX, Linux and/or Windows, you can run the following:

```
./build-bin [TAG_VERSION]
```

The resulting binaries will be placed in the `bin/` directory


### Run

To run Rebeat with debugging output enabled, run:

```
./rebeat -c rebeat.yml -e -d "*"
```


### Test

To test Rebeat, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/rebeat.template.json and etc/rebeat.asciidoc

```
make update
```


### Cleanup

To clean  Rebeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Rebeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/rswestmoreland
cd ${GOPATH}/github.com/rswestmoreland
git clone https://github.com/rswestmoreland/rebeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


