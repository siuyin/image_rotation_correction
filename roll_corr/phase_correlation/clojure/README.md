# Phase Correlation - Clojure Implementation

This is a high-performance implementation of the Phase Correlation approach for roll correction, leveraging native OpenCV and FFmpeg via JavaCV for real-time performance.

## Project Structure
- `src/roll_corr/phase_correlation/core.clj`: Main implementation using JavaCV.
- `test/roll_corr/phase_correlation/core_test.clj`: Test suite.
- `Makefile`: Automation for running tests and real-time processing.
- `deps.edn`: Dependency management (includes `javacv-platform`).

## How to Run
To process a video file in real-time (VIDEO_PATH is required):
```bash
clojure -M -m roll-corr.phase-correlation.core VIDEO_PATH -i 1000 -a 75 2>/dev/null
```

## Docker Containerization
To build and run as a container (run these commands from the project root):
```bash
# Build
docker build -t roll-corr-phase -f roll_corr/phase_correlation/clojure/Dockerfile .

# Run
docker run -e INTERVAL=500 -e AREA=50 -v /path/to/video:/app/video.mp4 roll-corr-phase /app/video.mp4
```
*Note: The build context is optimized using a root-level `.dockerignore` file, which excludes unnecessary files like `target/`, `.cpcache/`, and `.git/`.*

## Implementation Details
- **Real-Time Performance**: Utilizes native C++ OpenCV functions (`phaseCorrelate`, `warpPolar`) via JavaCV wrappers for high-speed computation.
- **Video Input**: Uses `FFmpegFrameGrabber` to directly extract and process frames from MP4 files.
- **Constraints Met**:
    - **Sampling**: Processes video at fixed time intervals (default 1000ms) using `setTimestamp`.
    - **Region of Interest**: Compares only the central $p\%$ (default 75%) of the image using `get-central-area`.
    - **Clean Output**: Redirect `2>/dev/null` to suppress verbose FFmpeg/JavaCV metadata.
