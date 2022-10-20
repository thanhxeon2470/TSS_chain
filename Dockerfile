FROM alpine:3.14


RUN set -x \
    && apk add --update --no-cache \
       ca-certificates gcompat \
    && rm -rf /var/cache/apk/* \
    && mkdir -p /opt/tss
WORKDIR /opt/tss

COPY TSS_chain /usr/local/bin/
COPY blockchain_genesis.db .
EXPOSE 3003 3000 8332
ENTRYPOINT ["TSS_chain"]