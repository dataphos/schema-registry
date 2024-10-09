# References the base image for Java 18 and maven
FROM maven:3.8.7-openjdk-18-slim AS build

# Maintainer Info
LABEL maintainer="Syntio Inc."

# Copy the source code to a new working directory
COPY registry/validity/external/validity-checker/src /home/app/src

# Copy the pom.xml to the root of the project
COPY registry/validity/external/validity-checker/pom.xml /home/app

COPY registry/licenses/validity-checker/LICENSE-3RD-PARTY.md /home/app/licenses/LICENSE-3RD-PARTY.md
COPY LICENSE /home/app/licenses/LICENSE

# Download dependecies and build
RUN mvn -f /home/app/pom.xml clean package

# References base image for Java 18 runtime
FROM openjdk:18-jdk-slim

# Copy the binaries in a new working directory
COPY --from=build /home/app/target/validity-checker.jar /home/checker/validity-checker.jar
COPY --from=build /home/app/licenses/LICENSE-3RD-PARTY.md /home/checker/licenses/LICENSE-3RD-PARTY.md
COPY --from=build /home/app/licenses/LICENSE /home/checker/licenses/LICENSE

# Expose port 8089 to the outside world
EXPOSE 8089

# change to a non-root user for security
RUN adduser --disabled-password --home /home/checker user
RUN chown -R user /home/checker
RUN chmod -R 500 /home/checker
USER user

# Set entrypoint of command that will run when container is started
ENTRYPOINT ["java","-jar","/home/checker/validity-checker.jar"]
