# Roll Correction

## Requirements

**Given** that I have a pan-tilt-zoom camera,  
and a mostly stationery scene  
and I capture an image with the camera  
and I define this image to be the reference image, R.  
and I define a central area of $p\%$ by area (defaulting to 75%) to be the region of interest for comparison.  

**When** the camera is rotated on any or all of its 3 axes, or has its lens zoom in or out, the camera will then capture a different but similar video sequence V, sampled every $m$ milliseconds (defaulting to 1000ms).  

**Then** I should get a report on how to rotate video frame V_n to be most similar to the reference image R, based on the defined region of interest.

## Plan

1. Sample video sequence V at interval $m$ (default 1000ms).
2. Crop or mask both the reference image R and video frame V_n to the central $p\%$ area (default 75%).
3. Determine amount of pitch, yaw and zoom for V_n.
4. Transform V_n correcting for the above such that an image that is highly similar, but rolled version of R.
5. Output the amount of rotation required to correct for the roll.
6. This roll correction must be streamed as each V_n is processed.

## Implementation Recommendations

### 1. Geometric Foundation
The relationship between the reference image $R$ and a video frame $V_n$ can be modeled as a **Homography** ($H$):
$$x_{V_n} \sim K_{V_n} \cdot \mathbf{R} \cdot K_{R}^{-1} \cdot x_{R}$$
Where $K$ represents internal camera parameters (zoom/focal length) and $\mathbf{R}$ is the 3D rotation matrix (Pitch, Yaw, Roll).

### 2. Fundamental Approaches

#### A. Feature-Based Estimation (Sparse)
- **Method**: Detect keypoints (landmarks), match them between images, and solve for the homography.
- **Pros**:
    - **Robust to Large Displacements**: Works even if the camera has moved significantly since the reference frame.
    - **Handles "Clutter" & Moving Objects**: Outliers (like a person walking through the frame) can be effectively filtered using RANSAC.
    - **Flexible Center of Rotation**: Does not assume rotation is centered on the image.
- **Cons**:
    - **Texture Dependent**: Requires scenes with distinct visual features; fails on blank walls or smooth surfaces.
    - **Variable Performance**: Processing time can fluctuate depending on the number of features detected in the scene.

#### B. Direct Intensity Alignment (Dense)
- **Method**: Minimize pixel intensity differences across the entire image using iterative optimization.
- **Pros**: High precision; utilizes all available data.
- **Cons**: Computationally intensive; sensitive to the initial guess.

#### C. Frequency Domain Analysis (Phase Correlation)
- **Method**: Use Fourier Transforms and Log-Polar mapping to convert rotation and scale into simple shifts.
- **Pros**:
    - **Deterministic Speed**: The fastest and most computationally inexpensive approach; execution time is constant regardless of scene content.
    - **Mathematically Elegant**: Handles rotation and scale changes as simple shifts in the frequency domain.
    - **Robust to Global Lighting Changes**: Operates primarily on the phase of the signal rather than raw intensity.
- **Cons**:
    - **Sensitive to Scene "Clutter"**: Moving objects or complex backgrounds can introduce noise that degrades the global alignment.
    - **Geometric Assumption**: Assumes rotation occurs around the image center (standard for most PTZ cameras, but a limitation if not).
    - **Aliasing Risks**: Highly periodic patterns (like a fence) can cause ambiguity in the matching.

### 3. Comparison & Efficiency Note
The implementation approaches are ranked below from fastest to most computationally expensive:

1.  **Frequency Domain Analysis (Phase Correlation)**: The most efficient starting point for high-performance streaming.
    - **Complexity**: $O(N \log N)$ via Fast Fourier Transform (FFT).
    - **Speed**: Single mathematical pass; no iteration required.
    - **Hardware-Friendly**: Highly optimized for modern CPUs/GPUs.
2.  **Feature-Based Estimation (Sparse)**: The next most efficient option; robust fallback.
    - **Complexity**: Scales with the number of keypoints (features), not total pixels.
    - **Speed**: Reduces data by processing only "points of interest"; uses fast binary descriptors.
3.  **Direct Intensity Alignment (Dense)**: The most computationally expensive.
    - **Complexity**: Scales with total pixels multiplied by optimization iterations.
    - **Speed**: Requires many iterative passes to converge; processes every pixel.

### 4. Streaming Strategy
To support real-time processing of $V_n$:
- **Temporal Prediction**: Use the solution from $V_{n-1}$ as the starting point for $V_n$.
- **Incremental Tracking**: Track changes between consecutive frames ($V_{n-1} \to V_n$) to maintain high throughput.
- **Drift Correction**: Periodically re-align against $R$ to eliminate accumulated error.
