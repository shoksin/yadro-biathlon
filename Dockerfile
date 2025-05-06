FROM ubuntu:latest
LABEL authors="shoksin"

ENTRYPOINT ["top", "-b"]