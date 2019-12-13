ARG GO_VERSION=1.11

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache  git

WORKDIR /src

COPY ./go.mod  ./
RUN go mod download


COPY . .

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /flanki_app .

FROM scratch AS final

COPY --from=builder /flanki_app /flanki_app
COPY /templates ./templates

EXPOSE 8080

ENTRYPOINT ["/flanki_app"]