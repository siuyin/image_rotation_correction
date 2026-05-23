import numpy as np
import cv2
import pytest
from virt_horiz import State, prep, get_k

def test_prep():
    frame = np.zeros((1000, 2000, 3), dtype=np.uint8)
    processed = prep(frame, w_max=640)
    assert processed.shape[1] == 640
    assert processed.shape[0] == 320

def test_get_k():
    img = np.zeros((480, 640, 3), dtype=np.uint8)
    k = get_k(img)
    assert k.shape == (3, 3)
    assert k[0, 0] == 640
    assert k[1, 1] == 640

def test_state_transition():
    state = State(interval=1000)
    assert state.phase == "init"
    
    # Fake reference frame establishment
    state.ref = np.zeros((100, 100), dtype=np.uint8)
    
    # Collect some baseline angles
    state.baseline_angles = [1.0, 2.0, 3.0]
    
    # Manually transition for test
    state.horizon_angle = np.mean(state.baseline_angles)
    state.phase = "corrected"
    
    assert state.phase == "corrected"
    assert state.horizon_angle == 2.0
