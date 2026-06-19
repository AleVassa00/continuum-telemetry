FROM ubuntu:latest
LABEL authors="aless"

ENTRYPOINT ["top", "-b"]