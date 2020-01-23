FROM alpine:3.7
COPY build/smug-linux-amd64 /smug-linux-amd64
COPY smug.yaml.template /smug.yaml
CMD /smug-linux-amd64 -config /smug.yaml
