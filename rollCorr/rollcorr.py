import cv2
import numpy as np
import argparse
import time

class State:
    def __init__(self, interval):
        self.ref = None
        self.last_timestamp = -interval
        self.last_angle = None
        self.interval_ms = interval

    def update(self, processed_frame, timestamp):
        if self.ref is None:
            self.ref = try_ref(processed_frame, self.ref, timestamp)
            return
        self.last_timestamp, self.last_angle = run_sample(
            self.ref, processed_frame, timestamp, self.last_timestamp,
            self.interval_ms, self.last_angle
        )

def orb_feats(ref_frame, src_frame):
    orb = cv2.ORB_create(nfeatures=1000)
    kp_ref, des_ref = orb.detectAndCompute(ref_frame, None)
    kp_src, des_src = orb.detectAndCompute(src_frame, None)
    if des_ref is None or des_src is None:
        return None
    return kp_ref, des_ref, kp_src, des_src

def match_pts(kp_ref, des_ref, kp_src, des_src):
    bf = cv2.BFMatcher(cv2.NORM_HAMMING, crossCheck=True)
    matches = sorted(bf.match(des_ref, des_src), key=lambda x: x.distance)
    if len(matches) < 8:
        return None, None
    p_ref = np.float32([kp_ref[m.queryIdx].pt for m in matches]).reshape(-1, 1, 2)
    p_src = np.float32([kp_src[m.trainIdx].pt for m in matches]).reshape(-1, 1, 2)
    return p_ref, p_src

def get_intrinsic_matrix(image):
    h, w = image.shape[:2]
    return np.array([[w, 0, w/2], [0, w, h/2], [0, 0, 1]], dtype=np.float32)

def get_roll_angle(ref_pts, src_pts, intrinsic_matrix):
    E, mask = cv2.findEssentialMat(src_pts, ref_pts, intrinsic_matrix, method=cv2.RANSAC, prob=0.999, threshold=1.0)
    _, R, _, _ = cv2.recoverPose(E, src_pts[mask.ravel()==1], ref_pts[mask.ravel()==1], intrinsic_matrix)
    return np.degrees(np.arctan2(R[2, 1], R[2, 2]))

def get_angle(ref_frame, src_frame):
    feats = orb_feats(ref_frame, src_frame)
    if not feats:
        return None
    ref_pts, src_pts = match_pts(*feats)
    if ref_pts is None:
        return None
    return get_roll_angle(ref_pts, src_pts, get_intrinsic_matrix(ref_frame))

def preprocess_frame(frame, max_width=640):
    h, w = frame.shape[:2]
    if w <= max_width:
        return frame
    scale = max_width / w
    return cv2.resize(frame, (max_width, int(h * scale)), interpolation=cv2.INTER_NEAREST)

def try_ref(proc_frame, ref_frame, timestamp):
    if ref_frame is None and np.mean(proc_frame) > 20:
        print(f"Ref at {timestamp:.1f}ms")
        return proc_frame
    return ref_frame

def apply_smoothing(new, last, alpha=0.3):
    if last is None:
        return new
    return alpha * new + (1 - alpha) * last

def run_sample(ref_frame, proc_frame, timestamp, last_ts, interval, last_ang):
    if timestamp - last_ts < interval:
        return last_ts, last_ang
    ang = get_angle(ref_frame, proc_frame)
    if ang is not None and abs(ang) < 45:
        if last_ang is None or abs(ang - last_ang) < 5.0:
            smoothed = apply_smoothing(ang, last_ang)
            print(f"{timestamp/1000.0:.1f} sec: {smoothed:.2f}")
            return timestamp, smoothed
    return last_ts, last_ang

def read_frame(capture, video_path):
    ret, frame = capture.read()
    if not ret:
        time.sleep(1)
        capture.open(video_path)
        return None
    return preprocess_frame(cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY))

def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("path")
    parser.add_argument("-m", "--interval", type=int, default=1000)
    return parser.parse_args()

def main():
    args = parse_args()
    cap = cv2.VideoCapture(args.path)
    state = State(args.interval)
    while True:
        proc = read_frame(cap, args.path)
        if proc is not None:
            state.update(proc, cap.get(cv2.CAP_PROP_POS_MSEC))

if __name__ == "__main__":
    main()
