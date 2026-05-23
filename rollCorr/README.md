# rollCorr

`rollCorr` is a Python-based tool designed to process video streams (files, RTSP, or HTTP) and calculate precise rotation (roll) correction angles against a reference frame in real-time.

## How It Works

- **Reference Frame Selection**: Automatically identifies the first valid (non-black) frame from the video stream to act as the ground truth reference.
- **Feature Matching**: Uses ORB feature detection and Essential Matrix estimation (RANSAC) to calculate the precise roll angle required to align incoming frames with the reference frame.
- **Real-time Streaming & Sampling**: Processes frames as they arrive from the stream, sampling at a user-configurable interval to output correction angles instantly, with temporal smoothing to filter out outliers.

## Requirements

- **Python**: 3.x
- **Dependency Management**: Uses `uv` for environment management.

## Usage

Run the tool by providing a path or URL to the video stream and an optional sampling interval:

```bash
uv run rollcorr.py <video_path_or_url> [-m <interval_ms>]
```

### Options
- `video_path_or_url`: Mandatory path to a local video file or network stream URL.
- `-m <interval_ms>`: Optional sampling interval in milliseconds (default: `1000`).

### Example

```bash
# Process a local video with a 1-second interval
uv run rollcorr.py ~/videos/my_video.mp4 -m 1000

# Process a network stream with a 500ms interval
uv run rollcorr.py rtsp://camera_ip:port/stream -m 500
```

## Running Local Tests

To run the unit tests included with the project:

```bash
uv run pytest
```
