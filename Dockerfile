
FROM golang:1.20

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY . ./
# https://andrei-calazans.com/posts/2021-06-23/passing-secrets-github-actions-docker
RUN --mount=type=secret,id=BOT_TOKEN \
    --mount=type=secret,id=OWNERID \
    export BOT_TOKEN=$(cat /run/secrets/BOT_TOKEN) && \
    export OWNERID=$(cat /run/secrets/OWNERID)

# Build
RUN go build -o /duder


# Run
CMD ["/duder"]