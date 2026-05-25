#include "feature_based.hpp"
#include <opencv2/features2d.hpp>
#include <opencv2/calib3d.hpp>
#include <cmath>

namespace RollCorrection {

cv::Mat get_central_area(const cv::Mat& image, double p_area) {
    if (image.empty()) return cv::Mat();
    
    int h = image.rows;
    int w = image.cols;
    double ratio = std::sqrt(p_area / 100.0);
    int new_w = static_cast<int>(w * ratio);
    int new_h = static_cast<int>(h * ratio);
    int x = (w - new_w) / 2;
    int y = (h - new_h) / 2;

    return image(cv::Rect(x, y, new_w, new_h)).clone();
}

void detect_and_describe(const cv::Mat& image, std::vector<cv::KeyPoint>& keypoints, cv::Mat& descriptors) {
    auto orb = cv::ORB::create();
    orb->detectAndCompute(image, cv::noArray(), keypoints, descriptors);
}

std::vector<cv::DMatch> match_features(const cv::Mat& descriptors_R, const cv::Mat& descriptors_V) {
    if (descriptors_R.empty() || descriptors_V.empty()) return {};
    
    auto matcher = cv::DescriptorMatcher::create(cv::DescriptorMatcher::BRUTEFORCE_HAMMING);
    std::vector<cv::DMatch> matches;
    matcher->match(descriptors_R, descriptors_V, matches);
    return matches;
}

cv::Mat solve_geometric_mapping(const std::vector<cv::KeyPoint>& kp_R, const std::vector<cv::KeyPoint>& kp_V, const std::vector<cv::DMatch>& matches) {
    if (matches.size() < 4) return cv::Mat();

    std::vector<cv::Point2f> src_pts, dst_pts;
    for (const auto& m : matches) {
        src_pts.push_back(kp_R[m.queryIdx].pt);
        dst_pts.push_back(kp_V[m.trainIdx].pt);
    }

    return cv::findHomography(src_pts, dst_pts, cv::RANSAC, 3.0);
}

TransformationParams decompose_transformation(const cv::Mat& mapping) {
    TransformationParams params = {0.0, 0.0, 1.0, 0.0};
    if (mapping.empty() || mapping.rows != 3 || mapping.cols != 3) return params;
    double h11 = mapping.at<double>(0, 0);
    double h21 = mapping.at<double>(1, 0);
    params.roll = std::atan2(h21, h11) * (180.0 / CV_PI);
    return params;
}

TransformationParams compute_frame_roll(const cv::Mat& frame, const std::vector<cv::KeyPoint>& ref_kp, const cv::Mat& ref_desc, double area_percent) {
    cv::Mat gray, central;
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    central = get_central_area(gray, area_percent);
    std::vector<cv::KeyPoint> kp;
    cv::Mat desc;
    detect_and_describe(central, kp, desc);
    return decompose_transformation(solve_geometric_mapping(ref_kp, kp, match_features(ref_desc, desc)));
}

void log_roll_correction(int frame_idx, double fps, double roll) {
    std::cout << "Time " << std::fixed << std::setprecision(2) << frame_idx / fps
              << "s - Required Roll Correction: " << roll << " degrees" << std::endl;
}

cv::VideoCapture open_video(const std::string& path) {
    cv::VideoCapture cap(path);
    if (!cap.isOpened()) throw std::runtime_error("Could not open video");
    return cap;
}

void initialize_reference(cv::VideoCapture& cap, double area_percent, std::vector<cv::KeyPoint>& ref_kp, cv::Mat& ref_desc) {
    cv::Mat frame, gray, central;
    if (!cap.read(frame)) throw std::runtime_error("Could not read first frame");
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    central = get_central_area(gray, area_percent);
    detect_and_describe(central, ref_kp, ref_desc);
}

void process_video(const std::string& path, int interval_ms, double area_percent) {
    auto cap = open_video(path);
    std::vector<cv::KeyPoint> ref_kp;
    cv::Mat ref_desc;
    initialize_reference(cap, area_percent, ref_kp, ref_desc);
    
    double fps = cap.get(cv::CAP_PROP_FPS);
    int step = std::max(1, static_cast<int>(interval_ms * fps / 1000.0));
    int idx = 0;
    cv::Mat frame;
    while (cap.read(frame)) {
        if (++idx % step != 0) continue;
        auto params = compute_frame_roll(frame, ref_kp, ref_desc, area_percent);
        log_roll_correction(idx, fps, params.roll);
    }
}
} // namespace RollCorrection
