
FROM golang:1.20

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY . ./

#Environment Variables
ENV BOT_TOKEN ${BOT_TOKEN}
ENV OWNERID ${OWNERID}

# Build
RUN go build -o /duder


# Run
CMD ["/duder"]