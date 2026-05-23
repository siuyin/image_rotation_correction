# levelHorizon

`levelHorizon` is a Go-based tool designed to calculate the rotation correction (roll) required to stabilize a video sequence.

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
go run main.go <path_to_video> [-m <interval_ms>]
```

### Options
- `<path_to_video>`: Mandatory path to the input video.
- `-m <interval_ms>`: Optional output interval in milliseconds (default: 1000ms).

### Example

```bash
go run main.go ../imageSequence/test.mp4 -m 500
```

## How It Works

The tool leverages FFmpeg's `vid.stab` library to perform motion analysis. It follows a two-pass process:

1.  **Pass 1: Motion Detection (`vidstabdetect`)**
    The tool runs `vidstabdetect` to calculate motion.

2.  **Pass 2: Transformation and Streaming (`vidstabtransform`)**
    The tool runs `vidstabtransform` in the background. It concurrently parses the generated `global_motions.trf` file to stream corrected rotation angles in real-time.

### Key Features
- **Smart Reference**: Automatically detects the first non-black frame (within the first 12 frames) and uses it as the reference point, rather than strictly defaulting to frame 0.
- **Time-based Output**: Configurable output frequency via the `-m` flag.
- **Real-time Streaming**: Results are streamed during the second pass, eliminating the need to wait for process completion.

## Implementation Details

- **Language**: Go (Standard Library + FFmpeg integration).
- **Core Logic**: Wraps FFmpeg command execution and incrementally parses transformation logs.
- **Precision**: Uses double-precision floating point for radian-to-degree conversion.

## Verification

The tool provides a clear mapping:

```text
Rotation Correction Results (relative to frame <ref_idx>):
Frame      Correction Angle (deg)
<frame>    <angle>              
...
```
