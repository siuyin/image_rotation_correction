# --- Build Stage ---
FROM ubuntu:24.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y build-essential cmake libopencv-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .
RUN mkdir build && cd build && cmake .. && make

# --- Final Stage ---
# Using gcr.io/distroless/cc-debian12 for a minimal footprint
FROM gcr.io/distroless/cc-debian12

WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/build/roll_corr /app/roll_corr

# Note: Distroless images do not contain shell tools, 
# so the ENTRYPOINT must be absolute.
ENTRYPOINT ["/app/roll_corr"]
