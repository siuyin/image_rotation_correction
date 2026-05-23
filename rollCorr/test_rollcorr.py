import numpy as np
import cv2
import pytest
from rollcorr import preprocess_frame, get_intrinsic_matrix

def test_prep():
    frame = np.zeros((1000, 2000, 3), dtype=np.uint8)
    processed = preprocess_frame(frame, max_width=640)
    assert processed.shape[1] == 640
    assert processed.shape[0] == 320

def test_get_k():
    img = np.zeros((480, 640, 3), dtype=np.uint8)
    k = get_intrinsic_matrix(img)
    assert k.shape == (3, 3)
    assert k[0, 0] == 640
    assert k[1, 1] == 640
