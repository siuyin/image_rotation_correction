# levelHorizon

`levelHorizon` is a Go-based tool designed to calculate the rotation correction (roll) required for each frame of a video sequence to align it with the first frame (frame 0) as a reference.

## Requirements

- **Go**: 1.26 or later.
- **FFmpeg**: Must be compiled with `libvidstab` support. You can verify this by running `ffmpeg -filters | grep vidstab`.

## Installation

1. Navigate to the `levelHorizon` directory:
   ```bash
   cd levelHorizon
   ```
2. Initialize and tidy the Go module (if not already done):
   ```bash
   go mod tidy
   ```

## Usage

Run the tool by providing the path to an input video:

```bash
go run main.go -i <path_to_video>
```

### Example

```bash
go run main.go -i ../imageSequence/test.mp4
```

## How It Works

The tool leverages FFmpeg's `vid.stab` library to perform high-quality motion analysis without requiring a full OpenCV/C++ environment. It follows a two-pass process:

1.  **Pass 1: Motion Detection (`vidstabdetect`)**
    The tool runs `vidstabdetect` with the `tripod=1` setting. This instructs the filter to treat the first frame as the absolute reference point and calculate the relative motion of all subsequent frames compared to it. The results are stored in a temporary `transforms.trf` file.

2.  **Pass 2: Global Transformation (`vidstabtransform`)**
    The tool then runs `vidstabtransform` using the previously generated transforms. It uses `tripod=true` to enable "virtual tripod" mode, which focuses on keeping the scene fixed relative to the reference. By setting `debug=true`, it generates a `global_motions.trf` file containing the precise rotation, translation, and zoom values for every frame.

3.  **Parsing and Conversion**
    The Go wrapper parses the `global_motions.trf` file, extracts the rotation (roll) value which is in radians, and converts it to degrees for the final output.

## Implementation Details

- **Language**: Go (Standard Library + FFmpeg integration).
- **Core Logic**: Wraps FFmpeg command execution and parses custom text-based transformation logs.
- **Precision**: Uses double-precision floating point for radian-to-degree conversion.

## Verification

The tool has been verified using `imageSequence/test.mp4`. The output format provides a clear mapping:

```text
Rotation Correction Results (relative to frame 0):
Frame      Correction Angle (deg)
0          0.0000              
1          0.2141              
2          0.4659              
...
```

The correction angle indicates how much the image needs to be rotated (clockwise or counter-clockwise) to "level" the horizon back to the state of the first frame.
