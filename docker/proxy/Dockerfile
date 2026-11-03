# The egress proxy (ADR 0002): tinyproxy that allows only the hosts in
# allowlist.conf. The image tag and digest are exact; tinyproxy is held to 1.11.3.
FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

RUN apk add --no-cache "tinyproxy~1.11.3"

COPY tinyproxy.conf allowlist.conf /etc/tinyproxy/

# Port 8888 is reached on the compose network only; it is never published.
USER tinyproxy
CMD ["tinyproxy", "-d", "-c", "/etc/tinyproxy/tinyproxy.conf"]
