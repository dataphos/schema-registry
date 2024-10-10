# Copyright 2024 Syntio Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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

COPY ./registry .
COPY LICENSE ./licenses/

RUN go mod tidy

RUN go build -buildvcs=false -o /app/sr ./cmd/janitorsr

FROM alpine:3.16

COPY --from=build /app/sr /app/sr
COPY --from=build /src/docs /app/docs
COPY --from=build /src/licenses/LICENSE-3RD-PARTY.md /app/licenses/
COPY --from=build /src/licenses/LICENSE /app/licenses/

# change to a non-root user for security
RUN adduser -D -h /app user
RUN chown -R user /app
RUN chmod -R 700 /app
USER user

EXPOSE 8080

CMD ["/app/sr"]
