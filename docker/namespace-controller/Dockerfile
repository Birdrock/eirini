ARG baseimage=scratch

FROM golang:1.13 as builder
WORKDIR /eirini/
COPY . .
RUN  CGO_ENABLED=0 GOOS=linux go build -mod vendor -trimpath -a -installsuffix cgo -o namespace-controller ./cmd/namespace-controller
ARG GIT_SHA
RUN if [ -z "$GIT_SHA" ]; then echo "GIT_SHA not set"; exit 1; else : ; fi

FROM ${baseimage}
COPY --from=builder /eirini/namespace-controller /usr/local/bin/namespace-controller
USER 1001
ENTRYPOINT [ "/usr/local/bin/namespace-controller", \
	"--config", \
	"/etc/eirini/config/namespace-controller.yml" \
]
ARG GIT_SHA
LABEL org.opencontainers.image.revision=$GIT_SHA \
      org.opencontainers.image.source=https://code.cloudfoundry.org/eirini
