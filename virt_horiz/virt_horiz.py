import cv2
import numpy as np
import argparse
import time

class State:
    def __init__(self, interval):
        self.ref = None
        self.interval_ms = interval
        self.phase = "init"
        self.baseline_angles = []
        self.horizon_angle = 0
        self.last_timestamp = -interval
        self.last_angle = None

    def update(self, proc, t):
        if self.ref is None:
            self.ref = try_ref(proc, self.ref, t)
            return
        self.last_timestamp, self.last_angle = run_sample(
            self.ref, proc, t, self.last_timestamp, self.interval_ms, self.last_angle, self
        )

def orb_feats(ref_frame, src_frame):
    orb = cv2.ORB_create(nfeatures=1000)
    kp_r, des_r = orb.detectAndCompute(ref_frame, None)
    kp_s, des_s = orb.detectAndCompute(src_frame, None)
    return (kp_r, des_r, kp_s, des_s) if des_r is not None and des_s is not None else None

def match_pts(kp_r, des_r, kp_s, des_s):
    bf = cv2.BFMatcher(cv2.NORM_HAMMING, crossCheck=True)
    matches = sorted(bf.match(des_r, des_s), key=lambda x: x.distance)
    if len(matches) < 8: return None, None
    p_r = np.float32([kp_r[m.queryIdx].pt for m in matches]).reshape(-1, 1, 2)
    p_s = np.float32([kp_s[m.trainIdx].pt for m in matches]).reshape(-1, 1, 2)
    return p_r, p_s

def get_k(img):
    h, w = img.shape[:2]
    return np.array([[w, 0, w/2], [0, w, h/2], [0, 0, 1]], dtype=np.float32)

def get_roll(p_r, p_s, k):
    E, mask = cv2.findEssentialMat(p_s, p_r, k, method=cv2.RANSAC, prob=0.999, threshold=1.0)
    _, R, _, _ = cv2.recoverPose(E, p_s[mask.ravel()==1], p_r[mask.ravel()==1], k)
    return np.degrees(np.arctan2(R[2, 1], R[2, 2]))

def get_angle(ref, src):
    f = orb_feats(ref, src)
    if not f: return None
    p_r, p_s = match_pts(*f)
    if p_r is None: return None
    return get_roll(p_r, p_s, get_k(ref))

def prep(frame, w_max=640):
    h, w = frame.shape[:2]
    if w <= w_max: return frame
    return cv2.resize(frame, (w_max, int(h * w_max / w)), interpolation=cv2.INTER_NEAREST)

def try_ref(proc, ref, t):
    if ref is None and np.mean(proc) > 20:
        print(f"Ref at {t:.1f}ms")
        return proc
    return ref

def apply_smoothing(new, last, alpha=0.3):
    return new if last is None else alpha * new + (1 - alpha) * last

def run_sample(ref_frame, proc_frame, timestamp, last_ts, interval, last_ang, state):
    if timestamp - last_ts < interval: return last_ts, last_ang
    ang = get_angle(ref_frame, proc_frame)
    if ang is None or abs(ang) >= 45: return last_ts, last_ang
    
    if state.phase == "init":
        if timestamp < 5000:
            state.baseline_angles.append(ang)
            print(f"{timestamp/1000.0:.1f} sec: Collecting ({len(state.baseline_angles)})")
            return timestamp, last_ang
        state.horizon_angle = np.mean(state.baseline_angles)
        state.phase = "corrected"
        print(f"Horizon established at {state.horizon_angle:.2f} deg")

    if last_ang is None or abs(ang - last_ang) < 5.0:
        smoothed = apply_smoothing(ang, last_ang)
        print(f"{timestamp/1000.0:.1f} sec: {smoothed - state.horizon_angle:.2f}")
        return timestamp, smoothed
    return last_ts, last_ang

def read_frame(capture, video_path):
    ret, frame = capture.read()
    if not ret:
        time.sleep(1); capture.open(video_path); return None
    return prep(cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY))

def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument("path"); p.add_argument("-m", "--interval", type=int, default=1000)
    return p.parse_args()

def main():
    args = parse_args(); cap = cv2.VideoCapture(args.path)
    state = State(args.interval)
    while True:
        proc = read_frame(cap, args.path)
        if proc is not None:
            state.update(proc, cap.get(cv2.CAP_PROP_POS_MSEC))

if __name__ == "__main__": main()
