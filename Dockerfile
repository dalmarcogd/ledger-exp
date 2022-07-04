FROM golang:1.18-buster as build-local
ENV GOOS=linux \
  GO111MODULE="on"

RUN mkdir -m 700 /root/.ssh; \
  touch -m 600 /root/.ssh/known_hosts; \
  ssh-keyscan github.com > /root/.ssh/known_hosts; \
  git config --global url."git@github.com:".insteadOf "https://github.com/"

WORKDIR /code

COPY go.mod .

# Download Go dependencies (include private modules)
#
# About this mount flag see in: https://docs.docker.com/develop/develop-images/build_enhancements/#using-ssh-to-access-private-data-in-builds
RUN --mount=type=ssh go mod download
