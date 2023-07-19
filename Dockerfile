# CONTAINER FOR BUILDING BINARY
FROM golang:1.17 AS build

# INSTALL DEPENDENCIES
RUN go install github.com/gobuffalo/packr/v2/packr2@v2.8.3
COPY go.mod go.sum /src/
RUN cd /src && go mod download

# BUILD BINARY
COPY . /src
RUN cd /src/db && packr2
RUN cd /src && make build

# CONTAINER FOR RUNNING BINARY
FROM alpine:3.16.0

# Install grpc health probe for being able to easily run health checks in docker
RUN wget https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.19/grpc_health_probe-linux-amd64 
RUN mv grpc_health_probe-linux-amd64 /usr/bin/grpc_health_probe
RUN chmod +x /usr/bin/grpc_health_probe

COPY --from=build /src/dist/zkevm-node /app/zkevm-node
COPY --from=build /src/config/environments/public/public.node.config.toml /app/example.config.toml

# PRELOAD CONFIGS
COPY ./test/aggregator.keystore /pk/aggregator.keystore
COPY ./test/config/test.genesis.config.json /app/genesis.json
COPY ./test/config/test.node.config.toml /app/config.toml
COPY ./test/sequencer.keystore /pk/sequencer.keystore

EXPOSE 8123
CMD ["/bin/sh", "-c", "/app/zkevm-node run"]
