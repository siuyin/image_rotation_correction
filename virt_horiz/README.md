# virt_horiz

`virt_horiz` processes video streams (files, RTSP, or HTTP) to establish an artificial horizon and output real-time rotation (roll) correction angles relative to that horizon.

## How It Works

- **Artificial Horizon Baseline**: Collects rotation correction angles over the first 5 seconds of the video stream to establish a baseline "horizon".
- **Real-time Leveling**: Calculates and outputs the precise roll angle required to level the camera relative to the established artificial horizon.
- **Robustness**: Includes temporal smoothing and outlier filtering for stable output.

## Requirements

- **Python**: 3.x
- **Dependency Management**: Uses `uv` for environment management.

## Usage

Run the tool by providing a path or URL to the video stream and an optional sampling interval:

```bash
uv run virt_horiz.py <video_path_or_url> [-m <interval_ms>]
```

### Example

```bash
uv run virt_horiz.py ~/tennis1.mp4 -m 1000
```
