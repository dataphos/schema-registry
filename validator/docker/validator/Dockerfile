FROM golang:1.21-alpine3.16 as build

LABEL maintainer="Syntio Inc."

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

RUN apk add --no-cache git

WORKDIR /src

COPY ./validator/go.mod ./validator/go.sum ./
RUN go mod download

COPY ./validator .
COPY LICENSE ./licenses/

RUN go mod tidy

RUN go build -buildvcs=false -o /app/validator ./cmd/centralconsumer

FROM alpine:3.16

COPY --from=build /app/validator /app/validator
COPY --from=build /src/config/validator.toml app/config/validator.toml
COPY --from=build /src/licenses/LICENSE-3RD-PARTY.md /app/licenses/
COPY --from=build /src/licenses/LICENSE /app/licenses/

# change to a non-root user for security
RUN adduser -D -h /app user
RUN chown -R user /app
RUN chmod -R 777 /app
USER user

ENTRYPOINT ["/app/validator", "-f", "/app/config/validator.toml" ]
