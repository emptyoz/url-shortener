FROM jenkins/jenkins:lts-jdk21

USER root

RUN apt-get update && apt-get install -y wget ca-certificates build-essential \
    && wget -q https://go.dev/dl/go1.25.3.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz \
    && ln -sf /usr/local/go/bin/go /usr/local/bin/go \
    && ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt \
    && rm -f go1.25.3.linux-amd64.tar.gz \
    && rm -rf /var/lib/apt/lists/*

USER jenkins
