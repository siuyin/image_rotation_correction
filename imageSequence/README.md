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
./imgseq [options] <path_to_video_file_or_stream_url>
```
The application supports local video files and live streams (RTSP, HTTP, HTTPS, RTMP). For RTSP streams, it automatically uses TCP transport for stability.
The extracted frames will be saved in the `output_frames/` directory.

### Options
*   `-w <int>`: Output image width (default: 640).
*   `-m <int>`: Step interval in milliseconds (default: 1000).

## Example
If you don't have a video file handy, you can generate a test video using FFmpeg:
```bash
ffmpeg -f lavfi -i testsrc=size=1280x720:rate=30 -t 5 test_video.mp4
```
Then run the application with custom width and interval:
```bash
./imgseq -w 320 -m 500 test_video.mp4
```
Expected output:
```text
Source: 1280x720, Target: 320x180, FPS: 2.00
at frame 0 (0.00s)
at frame 10 (5.00s)
Finished. Processed 10 frames.
done
```

## Testing with a Local Stream
To test the application's stream processing capabilities without an external IP camera, you can set up a local RTSP stream.

### 1. Start an RTSP Server
The easiest way is to use [mediamtx](https://github.com/bluenviron/mediamtx):
```bash
docker run --rm -it -e MTX_PROTOCOLS=tcp -p 8554:8554 bluenviron/mediamtx
```

### 2. Stream a file to the server
In another terminal, use FFmpeg to stream a video file (or a test source) to the server:
```bash
ffmpeg -re -f lavfi -i testsrc=size=1280x720:rate=30 -vcodec libx264 -preset ultrafast -tune zerolatency -f rtsp rtsp://localhost:8554/live
```

### 3. Run imgseq against the local stream
```bash
./imgseq rtsp://localhost:8554/live
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