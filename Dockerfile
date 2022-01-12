FROM alpine:3.14

ADD slash-milujipraci /usr/bin/
ADD Dockerfile /

ENTRYPOINT ["/usr/bin/slash-milujipraci"]

CMD ["--help"]
