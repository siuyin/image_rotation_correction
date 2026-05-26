# Build stage
FROM gocv/opencv:latest AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -o /rolldet_app main.go

# Prepare library directory
RUN mkdir -p /dist/usr/local/lib
RUN mkdir -p /dist/lib64
RUN cp /rolldet_app /dist/rolldet_app
RUN cp /lib64/ld-linux-x86-64.so.2 /dist/lib64/

# Dynamically find and copy ALL dependencies to /dist/usr/local/lib
RUN for lib in $(ldd /rolldet_app | grep "=> /" | awk '{print $3}'); do \
        cp "$lib" "/dist/usr/local/lib/"; \
    done

# Final stage
#FROM gcr.io/distroless/cc-debian12
FROM gcr.io/distroless/static-debian13:nonroot

# Copy everything from the dist directory
COPY --from=builder /dist /

# Set LD_LIBRARY_PATH to the flattened library directory
ENV LD_LIBRARY_PATH=/usr/local/lib

ENTRYPOINT ["/rolldet_app"]
