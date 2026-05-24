# Phase Correlation - Clojure Implementation

This is a fundamental implementation of the Phase Correlation approach for roll correction, using only the Clojure standard library and Java Math.

## Project Structure
- `src/roll_corr/phase_correlation/core.clj`: Main implementation.
- `test/roll_corr/phase_correlation/core_test.clj`: Test suite.
- `Makefile`: Automation for running tests.
- `deps.edn`: Dependency management.

## How to Run Tests
Execute the following command in this directory:
```bash
make test
```

## Implementation Details
- **Fundamentals Approach**: Functions are implemented from first principles.
- **Image Representation**: Images are represented as 2D vectors of intensity values (0.0 to 1.0).
- **Frequency Domain**: Placeholders for FFT/IFFT are provided; for production use, high-performance libraries like `neanderthal` or `opencv-clj` are recommended.
