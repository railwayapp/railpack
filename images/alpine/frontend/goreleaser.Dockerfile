FROM alpine

# The railpack binary is built during the GoReleaser process, which handles the
# cross-platform matrix build (OS/arch). GoReleaser then automatically copies
# the correct, pre-compiled binary into this image for each target architecture.
COPY railpack /

ENTRYPOINT ["/railpack", "frontend"]
