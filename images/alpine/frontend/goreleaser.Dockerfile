FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

# eMudhra fix: ca-certificates >= 20261020-r0.
RUN apk add --no-cache 'ca-certificates>=20261020-r0'

# The railpack binary is built during the GoReleaser process, which handles the
# cross-platform matrix build (OS/arch). GoReleaser then automatically copies
# the correct, pre-compiled binary into this image for each target architecture.
COPY railpack /

ENTRYPOINT ["/railpack", "frontend"]
