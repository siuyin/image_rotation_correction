# --- Build Stage ---
FROM ubuntu:24.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update
RUN apt-get install -y build-essential cmake libopencv-dev
RUN rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .
RUN mkdir build
RUN cd build && cmake ..
RUN cd build && make

# Collect dependencies
COPY collect_libs.sh /app/collect_libs.sh
RUN chmod +x /app/collect_libs.sh
RUN /app/collect_libs.sh

# --- Final Stage ---
FROM gcr.io/distroless/cc-debian12

# Copy the collected binary and libraries
COPY --from=builder /dist /

ENTRYPOINT ["/roll_corr"]
