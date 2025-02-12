# Declare build arguments at the very top
ARG RUNTIME=scratch-runtime

# Get certificates from Alpine (smaller than Ubuntu)
FROM alpine:latest@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099 AS certs
# No need to update here, we just use this for the certs
RUN apk add ca-certificates

# Used for everything except ftl-runner
FROM scratch AS scratch-runtime
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Used for ftl-runner
FROM ubuntu:24.04@sha256:72297848456d5d37d1262630108ab308d3e9ec7ed1c3286a32fe09856619a782 AS ubuntu-runtime
RUN apt-get update && apt-get install -y ca-certificates

# Final stage selection
FROM ${RUNTIME}
ARG EXTRA_FILES
ARG SERVICE

WORKDIR /home/ubuntu/
COPY . .


# Common environment variables
ENV PATH="$PATH:/home/ubuntu"

# Service-specific configurations
EXPOSE 8891
EXPOSE 8892
EXPOSE 8893

# Environment variables for all (most) services
ENV FTL_ENDPOINT="http://host.docker.internal:8892"
ENV FTL_BIND=http://0.0.0.0:8892
ENV FTL_ADVERTISE=http://127.0.0.1:8892
ENV FTL_DSN="postgres://host.docker.internal/ftl?sslmode=disable&user=postgres&password=secret"

# Controller-specific configurations
ENV FTL_CONTROLLER_CONSOLE_URL="*"

# Provisioner-specific configurations
ENV FTL_PROVISIONER_PLUGIN_CONFIG_FILE="/home/ubuntu/ftl-provisioner-config.toml"
USER 1000:1000
# Default command
CMD ["/home/ubuntu/svc"]
