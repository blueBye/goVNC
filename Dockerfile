FROM golang:1.22-alpine3.19

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and required files
COPY *.go ./
COPY recorder /app/recorder/
COPY client /app/client/

# Copy binaries
COPY ffmpeg/ ./
COPY vnc2video ./

# Build
RUN go build -o vnc_recorder

EXPOSE 8080

CMD ["/app/vnc_recorder"]
