FROM alpine

COPY railpack /

ENTRYPOINT ["railpack", "frontend"]
