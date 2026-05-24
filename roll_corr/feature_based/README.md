# Feature-Based Estimation (Sparse) Approach

This approach identifies and tracks distinct keypoints (landmarks) between the reference image and video frames to solve for the geometric transformation.

## Key Functions

### 1. `get_central_area(image, p_area)`
- **Input**: Image, percentage of central area (default 75%).
- **Output**: Cropped or masked image.
- **Description**: Extracts the central part of the image for comparison.

### 2. `detect_and_describe(image)`
- **Input**: Central part of the image.
- **Output**: List of key points and their unique signatures.
- **Description**: Identifies points of interest (e.g., corners, edges) and computes unique identifiers for each.

### 3. `match_features(signatures_R, signatures_V)`
- **Input**: Signatures from reference and video frame.
- **Output**: Set of matched point pairs.
- **Description**: Finds the most similar features between the two images.

### 4. `solve_geometric_mapping(matches)`
- **Input**: Matched point pairs.
- **Output**: 3x3 transformation matrix.
- **Description**: Uses a robust estimation method to find the best mathematical mapping between the images while ignoring incorrect matches.

### 5. `decompose_transformation(mapping)`
- **Input**: 3x3 transformation matrix.
- **Output**: Pitch, Yaw, Zoom, and Roll parameters.
- **Description**: Extracts the individual physical camera movements from the overall geometric mapping.

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
    signatures_R = detect_and_describe(R_central)
    
    # 2. Start Video Sampling Loop
    while camera_is_active():
        V_n = capture_video_frame()
        V_n_central = get_central_area(V_n, central_area_p)
        
        # 3. Find Key Points in Current Frame
        signatures_V = detect_and_describe(V_n_central)
        
        # 4. Match Points Between Reference and Current
        matches = match_features(signatures_R, signatures_V)
        
        # 5. Calculate Geometric Mapping
        mapping = solve_geometric_mapping(matches)
        
        # 6. Extract Camera Movement Parameters
        params = decompose_transformation(mapping)
        
        # 7. Report Roll Correction
        generate_report(params.roll)
        
        # 8. Wait for next sample
        sleep(sampling_interval_m)
```
