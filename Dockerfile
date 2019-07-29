# First stage: build the executable
FROM golang:1 AS builder

# Enable Go modules and vendor
ENV GO111MODULE=on GOFLAGS=-mod=vendor

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o rotator .

# Install chamber
ENV CHAMBER_VERSION=v2.1.0
RUN wget -q https://github.com//segmentio/chamber/releases/download/${CHAMBER_VERSION}/chamber-${CHAMBER_VERSION}-linux-amd64 -O /bin/chamber
RUN chmod +x /bin/chamber

# Final stage: the running container
FROM scratch AS final

WORKDIR /app
COPY --from=builder /app/rotator .
COPY --from=builder /bin/chamber /bin/chamber

CMD ["./rotator"]
