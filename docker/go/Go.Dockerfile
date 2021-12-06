# Stage 1: Create base build image with source code and Go modules.
FROM golang:1.16-alpine AS build_base
WORKDIR /go/src/ichor
COPY go.mod go.sum ./
RUN go mod download

# Stage 2: Build the service.
FROM build_base AS service_builder
COPY . /go/src/ichor
WORKDIR /go/src/ichor
RUN go build

# Stage 3: Create minimal image with service executable.
FROM golang:1.16-alpine AS service
WORKDIR /go/src/ichor
COPY --from=service_builder /go/src/ichor .
RUN mkdir -p data

EXPOSE 8080
ENV DISCORD_TOKEN ""

CMD ./ichor -t ${DISCORD_TOKEN} -a ${DEXCOM_ACCOUNT} -p ${DEXCOM_PASSWORD}