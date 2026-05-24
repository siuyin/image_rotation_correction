# Image Rotation Correction

This project provides high-performance, real-time solutions for detecting and correcting camera roll, pitch, and yaw for PTZ (Pan-Tilt-Zoom) cameras in mostly stationary scenes.

## Implemented Approaches

We have implemented two distinct approaches in Clojure, both optimized for real-time performance using native OpenCV bindings (via JavaCV):

1.  **[Phase Correlation Approach](roll_corr/phase_correlation/README.md)**: 
    - Uses Fourier Transforms and Log-Polar mapping.
    - Highly efficient, best for high-performance streaming.
2.  **[Feature-Based Estimation Approach](roll_corr/feature_based/README.md)**: 
    - Uses ORB feature detection and Homography estimation.
    - Robust against significant displacements and dynamic scene clutter.

## Project Structure
```text
.
├── roll_corr/
│   ├── phase_correlation/     # Phase Correlation implementation
│   └── feature_based/         # Feature-Based implementation
├── Dockerfile                 # (See subdirectory for approach-specific Dockers)
└── README.md                  # This file
```

## Running the Implementations

### Local Execution (Clojure)
Navigate to the implementation subdirectory (`phase_correlation` or `feature_based`):
```bash
cd roll_corr/<approach>/clojure
# Build the uberjar
clojure -T:build uber
# Run the application
java -jar target/app-1.0.0-standalone.jar /path/to/video.mp4 -i 1000 -a 75 2>/dev/null
```

### Docker Execution
From the project root:
```bash
# Build
docker build -t roll-corr-<approach> -f roll_corr/<approach>/clojure/Dockerfile .

# Run
docker run -e INTERVAL=1000 -e AREA=75 -v /path/to/video:/app/video.mp4 roll-corr-<approach> /app/video.mp4
```

## Documentation
For implementation-specific details, fundamental approaches, and performance characteristics, please refer to the READMEs located within each approach subdirectory.
