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

# References the base image for running Python applications
FROM python:3.7-alpine

# Maintainer Info
LABEL maintainer="Syntio Inc."

# Set the current working directory inside the container
WORKDIR /code

# Copy the dependencies file to the working directory inside the container
COPY validator/internal/validator/external/xml-validator/requirements.txt .

# Install dependencies from the file
RUN pip install --no-cache-dir -r requirements.txt

# Copy source (content of the local src directory) to the working directory inside the container
COPY validator/internal/validator/external/xml-validator .
COPY validator/licenses/xml-validator/LICENSE-3RD-PARTY.md app/licenses/LICENSE-3RD-PARTY.md
COPY LICENSE /app/licenses/LICENSE

# change to a non-root user for security
RUN adduser -D -h /code user
RUN chown -R user /code
RUN chmod -R 555 /code
USER user

EXPOSE 8081
CMD [ "python", "./main.py" ]
