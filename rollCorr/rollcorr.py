import cv2, numpy as np, argparse, sys, time

def orb_feats(ref, src):
    orb = cv2.ORB_create(nfeatures=1000)
    kp_r, des_r = orb.detectAndCompute(ref, None)
    kp_s, des_s = orb.detectAndCompute(src, None)
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

def try_ref(proc, ref, idx, time):
    if ref is None and np.mean(proc) > 20:
        print(f"Ref at {idx}, {time:.1f}ms")
        return proc
    return ref

def run_sample(ref, proc, time, last_t, interval, last_ang):
    if time - last_t < interval: return last_t, last_ang
    ang = get_angle(ref, proc)
    if ang is not None and abs(ang) < 45:
        if last_ang is None or abs(ang - last_ang) < 5.0:
            print(f"{time/1000.0:.1f} sec: {ang:.2f}")
            return time, ang
    return last_t, last_ang

def main():
    p = argparse.ArgumentParser()
    p.add_argument("path"); p.add_argument("-m", type=int, default=1000)
    a = p.parse_args(); cap = cv2.VideoCapture(a.path); ref, t_last, ang_last = None, -a.m, None
    while True:
        ret, f = cap.read()
        if not ret: time.sleep(1); cap.open(a.path); continue
        proc = prep(cv2.cvtColor(f, cv2.COLOR_BGR2GRAY))
        t = cap.get(cv2.CAP_PROP_POS_MSEC)
        if ref is None: ref = try_ref(proc, ref, 0, t) # Simplified index
        else: t_last, ang_last = run_sample(ref, proc, t, t_last, a.m, ang_last)

if __name__ == "__main__": main()
