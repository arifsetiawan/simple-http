FROM gcr.io/tetratelabs/tetrate-base:v0.1

ADD bin/simple-http /usr/local/bin/simple-http

ENTRYPOINT [ "/usr/local/bin/simple-http" ]