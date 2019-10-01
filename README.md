# Rabbitmq-beat

Welcome to Rabbitmq-beat.

Ensure that this folder is at the following location:
`${GOPATH}/src/hummer/rabbitmq-beat`

The Rabbitmq-beat is a elastic-beat, which listens to a user defined exchange on a Rabbit-MQ server. It consumes the messages which are send to this exchange and contain a routing-key, which matches with a user defined routing-key pattern. After the beat consumed a message it sends it to a user-defined output (e.g. elasticsearch).  

## Getting Started with Rabbitmq-beat

### Requirements

* [Golang](https://golang.org/dl/) 1.7

### Init Project
To get running with Rabbitmq-beat and also install the
dependencies, run the following command:

```
make setup
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

### Build

To build the binary for Rabbitmq-beat run the command below. This will generate a binary
in the same directory with the name rabbitmq-beat.

```
make
```

### Config

You can configurate the beat with the following variables in the `rabbit-mq.yml`:

| Variable              | Default value | Description                                                     |
|-----------------------|---------------|-----------------------------------------------------------------|
| rabbitmq_hostname     | "localhost"   | The hostname of the Rabbit-MQ server                            |
| rabbitmq_port         | "5672"        | The port under which the Rabbit-MQ server runs                  |
| rabbitmq_username     | ""            | The username of the user which connects to the Rabbit-MQ server |
| rabbitmq_passwd       | ""            | The password of the user which connects to the Rabbit-MQ server |
| rabbitmq_exchange     | ""            | The exchange to which the beats listens                         |
| rabbitmq_routing_keys | ["*.*"]       | The routing-key patterns the beat subscribes to                 |
| rabbitmq_log_config   | false         | Flag to log the config that start of the beat                   |


#### Server-sent event output
It is also possible to define a port for a server-sent event output in the in the `rabbit-mq.yml` with the `output.sse.port`(default: 8080) variable.

``` yml
output.sse.port: "8080"
```

### Run

To run Rabbitmq-beat with debugging output enabled, run:

``` bash
./rabbitmq-beat -c rabbitmq-beat.yml -e -d "*"
```

If you run the beat without the -c flag, the beat inits with the following values: 

### Docker

To pull the latest Rabbitmq-beat image, run:

``` bash
git pull geocode.igd.fraunhofer.de:4567/bfranke/rabbitmq-beat:latest
```

To start a container with the image, run:

``` bash
docker run -it geocode.igd.fraunhofer.de:4567/bfranke/rabbitmq-beat:latest
```

You can also use the `docker-compose.yml` to start a container:

``` bash
docker-compose up
```

#### Environment variables

You can configure the beat variables in a container with the following environment variables:

| Variable              | Environment variable      |
|-----------------------|---------------------------|
| rabbitmq_hostname     | RABBITMQ_HOSTNAME         | 
| rabbitmq_port         | RABBITMQ_PORT             |
| rabbitmq_username     | RABBITMQ_USERNAME         |
| rabbitmq_passwd       | RABBITMQ_PASSWD           | 
| rabbitmq_exchange     | RABBITMQ_EXCHANGE         | 
| rabbitmq_routing_keys | RABBITMQ_ROUTING_KEYS     |
| rabbitmq_log_config   | RABBITMQ_LOG_CONFIG       |                                     
| output.sse.port       | SSE_PORT                  |

### Cleanup

To clean  Rabbitmq-beat source code, run the following command:

```
make fmt
```

To clean up the build directory and generated artifacts, run:

```
make clean
```

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `fields.yml` by running the following command.

```
make update
```

### Test

To test Rabbitmq-beat, run the following command:

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

## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make release
```

This will fetch and create all images required for the build process. The whole process to finish can take several minutes.
