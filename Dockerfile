FROM golang:1.11 AS build
WORKDIR /src
ENV LAST_UPDATE=20180419
COPY . /src
RUN go get -d -v -t
RUN go test --cover ./...
RUN go build -v -tags netgo -o catnip

FROM alpine:3.8
ENV LAST_UPDATE=20180921
ENV DOCKER_API_VERSION=1.35
ENV TEMPLATE_ROOT="/srv/"
ENV EXTERNAL_HOSTNAME=""
LABEL authors="Joost van der Griendt <joostvdg@gmail.com>"
LABEL version="0.1.0"
LABEL description="Docker image for CATNIP"
CMD ["catnip"]
COPY --from=build /src/catnip /usr/local/bin/catnip
COPY index.html /srv/
RUN chmod +x /usr/local/bin/catnip
