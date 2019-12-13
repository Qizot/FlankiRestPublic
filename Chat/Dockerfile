ARG GO_VERSION=1.11

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache  git

WORKDIR /src

COPY ./go.mod  ./
RUN go mod download


COPY . .

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /chat .

FROM scratch AS final

COPY --from=builder /chat /chat

EXPOSE 8081

ENTRYPOINT ["/chat"]