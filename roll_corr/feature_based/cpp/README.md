re# C++ Feature-Based Roll Correction

This tool calculates the required roll correction for video frames using feature matching (ORB) and homography estimation, mirroring the functionality of the Clojure implementation.

## Prerequisites

To build and run this tool, you need:
- A C++17 compatible compiler (e.g., GCC 13+)
- CMake 3.10 or later
- OpenCV development files

### Installing Prerequisites (on Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y build-essential cmake libopencv-dev
```

## Build and Run

Use the provided `Makefile` to manage the project:

- **Build the project**:
  ```bash
  make build
  ```

- **Run memory leak check (requires valgrind)**:
  ```bash
  make leak-check VIDEO_FILE=/path/to/your/video.mp4
  ```

- **Build Docker image**:
  ```bash
  make docker-build
  ```

## Memory Analysis Note
Valgrind analysis shows "0 bytes definitely lost" in the application logic. Any "still reachable" memory reported is typical of initialized libraries (like OpenCV or Glibc) and does not indicate a memory leak in the tool itself.

## Usage

```bash
./roll_corr VIDEO_PATH [flags]
```

### Arguments:
- `VIDEO_PATH`: Path to video file (REQUIRED)
- `-i, --interval MS`: Sampling interval in ms (default: 1000)
- `-a, --area PERCENT`: Central area percentage (default: 75)
- `-h, --help`: Print this help

## Core Algorithm
1. **Feature Detection**: Uses `cv::ORB` to detect keypoints and compute descriptors.
2. **Matching**: Uses `cv::BFMatcher` with Hamming distance to match features between the reference and target frames.
3. **Geometric Mapping**: Uses `cv::findHomography` with RANSAC to estimate the transformation.
4. **Decomposition**: Extracts the roll angle by decomposing the homography matrix.
