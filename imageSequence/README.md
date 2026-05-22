# Generate an image sequence from a video stream

This application extracts frames from a video file, resizes them to a standard width, and saves them as a sequence of JPEG images.

## Pre-requisites
```bash
sudo apt-get install ffmpeg
```

## Usage
### 1. Build the application
```bash
go build -o imgseq main.go
```

### 2. Run the application
```bash
./imgseq <path_to_video_file>
```
The extracted frames will be saved in the `output_frames/` directory.

## Example
If you don't have a video file handy, you can generate a test video using FFmpeg:
```bash
ffmpeg -f lavfi -i testsrc=size=1280x720:rate=30 -t 5 test_video.mp4
```
Then run the application:
```bash
./imgseq test_video.mp4
```
Expected output:
```text
Source: 1280x720, Target: 640x360, FPS: 30.00
at 100 frames...
at 150 frames...
Finished. Processed 150 frames.
done
```

## Technical Analysis
...
The application is written in Go and uses the `ffmpeg` and `ffprobe` command-line tools as subprocesses for video processing. This approach avoids the complexities and versioning issues associated with CGO bindings.

### Workflow
1.  **Metadata Extraction**: Uses `ffprobe` to determine the input video's width, height, and average frame rate.
2.  **Target Dimension Calculation**: Automatically calculates the target height based on a fixed `targetWidth` (640px), ensuring the original aspect ratio is maintained and the height is an even number.
3.  **Frame Streaming**: Launches an `ffmpeg` subprocess to:
    *   Decode the video.
    *   Scale it to the target resolution.
    *   Stream raw `RGB24` frame data to `stdout`.
4.  **Processing Loop**:
    *   Reads exact frame-sized chunks from the `ffmpeg` pipe.
    *   Converts raw bytes into Go-native `image.RGBA` objects.
5.  **Output**: Saves each frame as a JPEG image in the `output_frames/` directory. Filenames include the timestamp offset (e.g., `frame_offset_0000.123.jpg`) calculated from the frame index and frame rate.

### Key Features
*   **CGO-Free**: No external C libraries are linked at compile time, making the binary highly portable across systems with `ffmpeg` installed.
*   **Aspect-Ratio Preservation**: Correctly handles scaling for any input resolution.
*   **Sequential Naming**: Uses precise timestamp-based naming for easy reconstruction of the sequence.