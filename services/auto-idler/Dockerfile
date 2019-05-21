# build the binary
FROM golang AS builder

# bring in all the packages
COPY . /go/src/github.com/amazeeio/lagoon/services/autoidler/

WORKDIR /go/src/github.com/amazeeio/lagoon/services/autoidler/

# get any imports as required
RUN set -x && go get -v .
# compile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o auto-idler .

# put the binary into container
# use the commons image
FROM ${IMAGE_REPO:-lagoon}/commons

WORKDIR /app/

# bring the auto-idler binary from the builder
COPY --from=builder /go/src/github.com/amazeeio/lagoon/services/autoidler/auto-idler .
# bring the cron scripts in, but they just run curl
COPY idle-clis.sh .
COPY idle-services.sh .

ENV LAGOON=auto-idler

# set defaults (also needs LOGSDB_ADMIN_PASSWORD)
ENV JWTSECRET=super-secret-string \
    JWTAUDIENCE=api.dev \
    PROJECT_REGEX=".+" \
    HTTP_PORT="3000" \
    LOGSDB_ADMIN_HOST="http://logs-db:9200" \
    GRAPHQL_ENDPOINT="http://api:3000/graphql" \
    SERVICE_IDLER_TIME="4h"

ENTRYPOINT ["/sbin/tini", "--", "/lagoon/entrypoints.sh"]
CMD ["/app/auto-idler"]
