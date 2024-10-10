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

# References the base image for Java 18 and maven
FROM maven:3.8.7-openjdk-18-slim AS build

# Maintainer Info
LABEL maintainer="Syntio Inc."

# Copy the source code to a new working directory
COPY registry/compatibility/external/compatibility-checker/src /home/app/src

# Copy the pom.xml to the root of the project
COPY registry/compatibility/external/compatibility-checker/pom.xml /home/app

COPY registry/licenses/compatibility-checker/LICENSE-3RD-PARTY.md /home/app/licenses/LICENSE-3RD-PARTY.md
COPY LICENSE /home/app/licenses/LICENSE

# Download dependecies and build
RUN mvn -f /home/app/pom.xml clean package

# References base image for Java 18 runtime
FROM openjdk:18-jdk-slim

# Copy the binaries in a new working directory
COPY --from=build /home/app/target/compatibility-checker.jar /home/checker/compatibility-checker.jar
COPY --from=build /home/app/licenses/LICENSE-3RD-PARTY.md /home/checker/licenses/LICENSE-3RD-PARTY.md
COPY --from=build /home/app/licenses/LICENSE /home/checker/licenses/LICENSE

# Expose port 8088 to the outside world
EXPOSE 8088

# change to a non-root user for security
RUN adduser --disabled-password --home /home/checker user
RUN chown -R user /home/checker
RUN chmod -R 500 /home/checker
USER user

# Set entrypoint of command that will run when container is started
ENTRYPOINT ["java","-jar","/home/checker/compatibility-checker.jar"]
