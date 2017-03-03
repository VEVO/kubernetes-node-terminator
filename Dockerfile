FROM scratch

ADD terminator /terminator

ENTRYPOINT ["/terminator"]
