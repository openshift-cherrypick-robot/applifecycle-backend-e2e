# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

RUN  microdnf update -y \
        && rpm -e --nodeps tzdata \
        && microdnf install tzdata \
        && microdnf install openssh-clients \
        && microdnf install curl \
        && microdnf install procps \
        && microdnf install tar \
        && microdnf clean all

ENV USER_UID=1001 \
    USER_NAME=app-backend \
    ZONEINFO=/usr/share/timezone

COPY COMPONENT_VERSION /COMPONENT_VERSION

RUN export COMPONENT_VERSION=$(cat /COMPONENT_VERSION); git clone -b release-${COMPONENT_VERSION} --single-branch https://github.com/open-cluster-management/applifecycle-backend-e2e.git /opt/e2e

WORKDIR /opt/e2e/client/canary

# the test data is in the binary format
ENTRYPOINT go test -v -timeout 30m

# Document that the service listens on port 8765.
EXPOSE 8765
