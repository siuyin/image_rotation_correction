# Feature-Based Estimation - Clojure Implementation

This is a high-performance implementation of the Feature-Based approach for roll correction, leveraging native OpenCV and FFmpeg via JavaCV for real-time performance.

## Project Structure
- `src/roll_corr/feature_based/core.clj`: Main implementation using JavaCV (ORB + Homography).
- `test/roll_corr/feature_based/core_test.clj`: Test suite.
- `Makefile`: Automation for running tests and real-time processing.
- `deps.edn`: Dependency management (includes `javacv-platform`).

## How to Run
To process a video file in real-time (defaulting to `~/tennis1.mp4`):
```bash
make run VIDEO=/path/to/your/video.mp4
```

## Implementation Details
- **Real-Time Performance**: Utilizes native C++ OpenCV functions (`ORB`, `findHomography`) via JavaCV wrappers for high-speed tracking and geometric mapping.
- **Video Input**: Uses `FFmpegFrameGrabber` to directly extract and process frames from MP4 files.
- **Constraints Met**:
    - **Sampling**: Processes video at fixed time intervals (default 1000ms) using `setTimestamp`.
    - **Region of Interest**: Compares only the central $p\%$ (default 75%) of the image.
