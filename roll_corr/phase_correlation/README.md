# Phase Correlation (Frequency Domain) Approach

This approach utilizes the Fourier Transform to identify the transformation between the reference image and video frames in the frequency domain.

## Key Functions

### 1. `get_central_area(image, p_area)`
- **Input**: Image, percentage of central area (default 75%).
- **Output**: Cropped or masked image.
- **Description**: Extracts the central part of the image for comparison to reduce edge noise and focus on the subject.

### 2. `to_log_polar(image)`
- **Input**: Image in standard coordinates.
- **Output**: Image in Log-Polar coordinates.
- **Description**: Maps the image such that rotations and scaling (zoom) become linear translations.

### 3. `compute_phase_correlation(img1, img2)`
- **Input**: Two images (or their Fourier transforms).
- **Output**: Shift vector (relative change).
- **Description**: Computes the cross-power spectrum to find the peak representing the global shift between images.

### 4. `estimate_pitch_yaw(R_central, V_n_central)`
- **Input**: Reference central area, Video Frame central area.
- **Output**: Pitch and Yaw differences.
- **Description**: Uses phase correlation in the standard coordinate system to find how much the camera tilted or panned.

### 5. `estimate_zoom_roll(R_central, V_n_central)`
- **Input**: Reference central area, Video Frame central area.
- **Output**: Zoom ratio and Roll angle.
- **Description**: Uses phase correlation in the Log-Polar coordinate system to identify scale and rotation changes.

### 6. `generate_report(roll_angle)`
- **Input**: Calculated roll correction.
- **Output**: Formatted report.
- **Description**: Streams the required rotation to the user.

## Main Function Pseudocode

```python
def main(reference_image_path, sampling_interval_m=1000, central_area_p=75):
    # 1. Initialize Reference
    R = load_image(reference_image_path)
    R_central = get_central_area(R, central_area_p)
    
    # 2. Start Video Sampling Loop
    while camera_is_active():
        V_n = capture_video_frame()
        V_n_central = get_central_area(V_n, central_area_p)
        
        # 3. Estimate Pitch and Yaw (Tilt/Pan)
        pitch, yaw = estimate_pitch_yaw(R_central, V_n_central)
        
        # 4. Estimate Zoom and Roll
        zoom, roll = estimate_zoom_roll(R_central, V_n_central)
        
        # 5. Report Roll Correction
        generate_report(roll)
        
        # 6. Wait for next sample
        sleep(sampling_interval_m)
```
