# References the base image for Java 11 and maven
FROM maven:3.8.4-openjdk-17-slim AS build

# Maintainer Info
LABEL maintainer="Syntio Inc."

# Copy the source code to a new working directory
COPY validator/internal/validator/external/csv-validator/src /home/app/src

# Copy the pom.xml to the root of the project
COPY validator/internal/validator/external/csv-validator/pom.xml /home/app

COPY validator/licenses/csv-validator/LICENSE-3RD-PARTY.md /home/app/licenses/LICENSE-3RD-PARTY.md
COPY LICENSE /home/app/licenses/LICENSE

# Download dependecies and build
RUN mvn -f /home/app/pom.xml clean package

# References base image for Java 11 runtime
FROM openjdk:22-ea-17-jdk-slim

# Copy the binaries in a new working directory
COPY --from=build /home/app/target/csv-validator-0.0.1-SNAPSHOT.jar /home/csv/validator.jar
COPY --from=build /home/app/licenses/LICENSE-3RD-PARTY.md /home/csv/licenses/LICENSE-3RD-PARTY.md
COPY --from=build /home/app/licenses/LICENSE /home/csv/licenses/LICENSE

# Expose port 8080 to the outside world
EXPOSE 8080

# change to a non-root user for security
RUN adduser --disabled-password --home /home/csv user
RUN chown -R user /home/csv
RUN chmod -R 555 /home/csv
USER user

# Set entrypoint of command that will run when container is started
ENTRYPOINT ["java","-jar","/home/csv/validator.jar"]

# From the helper-functions directory:
# sudo docker build -t centralconsumer-csv-val .
# sudo docker run centralconsumer-csv-val
