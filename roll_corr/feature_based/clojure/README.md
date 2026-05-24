# Feature-Based Estimation - Clojure Implementation

This is a fundamental implementation of the Feature-Based Estimation approach for roll correction, using only the Clojure standard library and Java Math.

## Project Structure
- `src/roll_corr/feature_based/core.clj`: Main implementation.
- `test/roll_corr/feature_based/core_test.clj`: Test suite.
- `Makefile`: Automation for running tests.
- `deps.edn`: Dependency management.

## How to Run Tests
Execute the following command in this directory:
```bash
make test
```

## Implementation Details
- **Fundamentals Approach**: Functions are implemented from first principles.
- **Key Point Detection**: Uses a simplified threshold-based detector.
- **Matching**: Implements a nearest-neighbor signature matcher.
- **Geometric Mapping**: Placeholders for robust Homography solvers are provided; for production use, libraries like `opencv-clj` or `origami` are recommended.
