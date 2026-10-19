FROM registry.suse.com/bci/bci-base:15.7
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETOS

ENV TARGETPLATFORM=${TARGETPLATFORM:-"linux/amd64"} ARCH=${TARGETARCH:-"amd64"} OS=${TARGETOS:-"linux"}

RUN zypper -n install curl ca-certificates tar gzip
RUN mkdir /home/shell && \
    echo '. /etc/profile.d/bash_completion.sh' >> /home/shell/.bashrc && \
    echo 'PS1="> "' >> /home/shell/.bashrc

ENV AUTOK3S_CONFIG=/root/.autok3s
ENV DOCKER_HOST=unix:///var/run/docker.sock
ENV HOME=/root

WORKDIR /home/shell
VOLUME /root/.autok3s
COPY bin/${TARGETPLATFORM}/autok3s /usr/local/bin/autok3s
ENTRYPOINT ["autok3s"]
CMD ["serve", "--bind-address=0.0.0.0"]
