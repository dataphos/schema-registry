FROM golang:alpine3.15 AS build

LABEL maintainer="Syntio Inc."

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

RUN apk add --no-cache git

WORKDIR /src

COPY ./registry/go.mod ./registry/go.sum ./
RUN go mod download

COPY ./schema-registry .
COPY LICENSE ./licenses/

RUN go mod tidy

RUN go build -buildvcs=false -o /app/initdb ./cmd/initdb

FROM alpine:3.16

COPY --from=build /app/initdb /app/initdb
COPY --from=build /src/licenses/LICENSE-3RD-PARTY.md /app/licenses/
COPY --from=build /src/licenses/LICENSE /app/licenses/

# change to a non-root user for security
RUN adduser -D -h /app user
RUN chown -R user /app
RUN chmod -R 700 /app
USER user

CMD ["/app/initdb"]
