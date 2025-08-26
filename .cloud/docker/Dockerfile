# syntax=docker/dockerfile:1

ARG KIND=plugins
ARG USERNAME=codeclarity
ARG PLUGINNAME=plugin

# DEV IMAGE
FROM golang:1.24.5 AS plugin-dev
ARG KIND
ARG PLUGINNAME
ENV PATH="$PATH:/codeql"
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root
RUN apt update -y && apt upgrade -y
RUN apt install -y zstd nodejs
RUN curl -L https://github.com/github/codeql-action/releases/download/codeql-bundle-v2.21.1/codeql-bundle-linux64.tar.zst > codeql-bundle.tar.zst
RUN tar --zstd -xvf codeql-bundle.tar.zst
RUN mv codeql /codeql
RUN rm codeql-bundle.tar.zst
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air.toml"]

# DEBUG IMAGE
FROM golang:1.24.5 AS plugin-debug
ARG KIND
ARG PLUGINNAME
ENV PATH="$PATH:/codeql"
ENV KIND=${KIND}
ENV PLUGINNAME=${PLUGINNAME}
RUN apt update -y && apt upgrade -y
RUN apt install -y zstd nodejs
RUN curl -L https://github.com/github/codeql-action/releases/download/codeql-bundle-v2.21.1/codeql-bundle-linux64.tar.zst > codeql-bundle.tar.zst
RUN tar --zstd -xvf codeql-bundle.tar.zst
RUN mv codeql /codeql
RUN rm codeql-bundle.tar.zst
RUN go install github.com/go-delve/delve/cmd/dlv@latest
ADD .cloud/docker/config/* /root
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
CMD ["/go/bin/dlv", "debug", ".", "--headless", "--listen=:40000", "--accept-multiclient"]
# CMD tail -f /dev/null


# BUILD IMAGE
FROM golang:1.24.5 AS plugin-build
ARG KIND
ARG PLUGINNAME
COPY . /codeclarity/${KIND}/${PLUGINNAME}
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root
RUN rm go.work
RUN apt update -y && apt upgrade -y
RUN apt install git
RUN [ "go", "build", "-o", "plugin" ]

# PRODUCTION IMAGE
FROM debian:stable-slim AS plugin
ARG KIND
ARG USERNAME
ARG PLUGINNAME

RUN apt-get update -y && apt-get upgrade -y
RUN apt-get install -y zstd nodejs curl
# Clean cache
RUN apt-get clean
RUN rm -rf /var/lib/apt/lists/*

RUN groupadd -g 1001 ${USERNAME}
RUN useradd --uid 1001 --gid 1001 --create-home ${USERNAME}
USER ${USERNAME}

RUN mkdir /home/${USERNAME}/bin
WORKDIR /home/${USERNAME}/bin
RUN curl -L https://github.com/github/codeql-action/releases/download/codeql-bundle-v2.21.1/codeql-bundle-linux64.tar.zst > codeql-bundle.tar.zst
RUN tar --zstd -xvf codeql-bundle.tar.zst
RUN rm codeql-bundle.tar.zst
ENV PATH="$PATH:/home/$USERNAME/bin/codeql"

WORKDIR /home/${USERNAME}/${PLUGINNAME}
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/${KIND}/${PLUGINNAME}/plugin .
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/${KIND}/${PLUGINNAME}/config.json .

ENTRYPOINT [ "./plugin" ]
