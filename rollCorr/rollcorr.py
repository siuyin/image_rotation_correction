import cv2
import numpy as np
import argparse
import sys
import time

def extract_orb_features(img_ref, img_src):
    orb = cv2.ORB_create(nfeatures=3000)
    kp_ref, des_ref = orb.detectAndCompute(img_ref, None)
    kp_src, des_src = orb.detectAndCompute(img_src, None)
    if des_ref is None or des_src is None:
        return None
    return kp_ref, des_ref, kp_src, des_src

def match_feature_points(kp_ref, des_ref, kp_src, des_src):
    bf = cv2.BFMatcher(cv2.NORM_HAMMING, crossCheck=True)
    matches = sorted(bf.match(des_ref, des_src), key=lambda x: x.distance)
    if len(matches) < 8:
        return None, None
    pts_ref = np.float32([kp_ref[m.queryIdx].pt for m in matches]).reshape(-1, 1, 2)
    pts_src = np.float32([kp_src[m.trainIdx].pt for m in matches]).reshape(-1, 1, 2)
    return pts_ref, pts_src

def estimate_intrinsic_matrix(img):
    h, w = img.shape[:2]
    focal_length = w
    center_x, center_y = w / 2, h / 2
    return np.array([[focal_length, 0, center_x], [0, focal_length, center_y], [0, 0, 1]], dtype=np.float32)

def extract_roll_correction(pts_ref, pts_src, K):
    E, mask = cv2.findEssentialMat(pts_src, pts_ref, K, method=cv2.RANSAC, prob=0.999, threshold=1.0)
    inliers_src = pts_src[mask.ravel() == 1]
    inliers_ref = pts_ref[mask.ravel() == 1]
    _, R, _, _ = cv2.recoverPose(E, inliers_src, inliers_ref, K)
    return np.degrees(np.arctan2(R[2, 1], R[2, 2]))

def get_correction_angle(img_ref, img_src):
    features = extract_orb_features(img_ref, img_src)
    if features is None: return None
    pts_r, pts_s = match_feature_points(*features)
    if pts_r is None: return None
    K = estimate_intrinsic_matrix(img_ref)
    return extract_roll_correction(pts_r, pts_s, K)

def main():
    parser = argparse.ArgumentParser(description="Calculate roll correction from video stream.")
    parser.add_argument("video_path", help="Path or URL to video")
    parser.add_argument("-m", type=int, default=1000, help="Interval in ms")
    args = parser.parse_args()

    cap = cv2.VideoCapture(args.video_path)
    fps = cap.get(cv2.CAP_PROP_FPS) or 30
    frame_interval = int(fps * args.m / 1000.0)
    print(f"FPS: {fps:.2f}, Calculated frame interval: {frame_interval}")
    
    img_ref = None
    frame_idx = 0
    last_sample_time = -args.m 
    
    while True:
        ret, frame = cap.read()
        if not ret:
            time.sleep(1)
            cap.open(args.video_path)
            continue
            
        current_msec = cap.get(cv2.CAP_PROP_POS_MSEC)
        gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
        
        if img_ref is None:
            if np.mean(gray) > 20:
                img_ref = gray
                print(f"Reference frame established at index {frame_idx}, time {current_msec:.1f}ms")
            frame_idx += 1
            continue
            
        if current_msec - last_sample_time >= args.m:
            angle = get_correction_angle(img_ref, gray)
            if angle is not None:
                print(f"{current_msec/1000.0:.1f} sec: {angle:.2f}")
                last_sample_time = current_msec
        
        frame_idx += 1

if __name__ == "__main__":
    main()
