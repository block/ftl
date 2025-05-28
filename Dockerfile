# Declare build arguments at the very top
ARG RUNTIME=scratch-runtime

# Get certificates from Alpine (smaller than Ubuntu)
FROM alpine:latest@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c AS certs
# No need to update here, we just use this for the certs
RUN apk add ca-certificates

# Used for everything except ftl-runner
FROM scratch AS scratch-runtime
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Used for ftl-runner
FROM ubuntu:24.04@sha256:6015f66923d7afbc53558d7ccffd325d43b4e249f41a6e93eef074c9505d2233 AS ubuntu-runtime
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
EXPOSE 8892

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
