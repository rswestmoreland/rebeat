# Rebeat

## Description

This application is used as a Relay/Repeater/Receiver (hence the name "Re"beat) for  messages sent by other Elastic Beats via Lumberjack v2 (Beats Protocol) which is used for Logstash destinations.

Ensure that this folder is at the following location:
`${GOPATH}/github.com/rswestmoreland`

## Getting Started with Rebeat

### Configuration Options

- `rebeat.address` : The address on which the process will listen (Default: 127.0.0.1)
- `rebeat.port` : The port on which the process will listen (Default = 5044)
- `rebeat.timeout` : Number of seconds to wait for data before closing connection (Default = 0 for no timeout)
- `tls.enable` : Enable optional TLS support (Default = false)
- `tls.certification` : Specify path to server's tls cert (pem or crt format)
- `tls.key` : Specify path to server's tls key

#### Considerations

This project is *new* and will require more updates before it is officially production ready.  


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


### Clone

To clone Rebeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/rswestmoreland
cd ${GOPATH}/github.com/rswestmoreland
git clone https://github.com/rswestmoreland/rebeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


