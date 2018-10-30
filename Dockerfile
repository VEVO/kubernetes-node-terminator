# Getting a small image with only the binary
FROM scratch
COPY terminator /
ENTRYPOINT ["/terminator"]
