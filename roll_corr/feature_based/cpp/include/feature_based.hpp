#ifndef FEATURE_BASED_HPP
#define FEATURE_BASED_HPP

#include <opencv2/opencv.hpp>
#include <vector>

namespace RollCorrection {

/**
 * @brief Extracts the central area of an image.
 * @param image The input image.
 * @param p_area The percentage of the area to extract (0-100).
 * @return The cropped central area.
 */
cv::Mat get_central_area(const cv::Mat& image, double p_area);

/**
 * @brief Detects and describes ORB features in an image.
 * @param image The input image.
 * @param keypoints Vector to store detected keypoints.
 * @param descriptors Matrix to store feature descriptors.
 */
void detect_and_describe(const cv::Mat& image, std::vector<cv::KeyPoint>& keypoints, cv::Mat& descriptors);

/**
 * @brief Matches features between two sets of descriptors.
 * @param descriptors_R Reference descriptors.
 * @param descriptors_V Target descriptors.
 * @return Vector of matches.
 */
std::vector<cv::DMatch> match_features(const cv::Mat& descriptors_R, const cv::Mat& descriptors_V);

/**
 * @brief Solves for the geometric mapping (homography) between two sets of keypoints.
 * @param kp_R Reference keypoints.
 * @param kp_V Target keypoints.
 * @param matches Matches between reference and target.
 * @return The homography matrix.
 */
cv::Mat solve_geometric_mapping(const std::vector<cv::KeyPoint>& kp_R, const std::vector<cv::KeyPoint>& kp_V, const std::vector<cv::DMatch>& matches);

/**
 * @brief Decomposes a transformation matrix to extract camera parameters.
 * @param mapping The transformation matrix (3x3 homography).
 * @return A struct containing pitch, yaw, zoom, and roll.
 */
struct TransformationParams {
    double pitch;
    double yaw;
    double zoom;
    double roll;
};

TransformationParams decompose_transformation(const cv::Mat& mapping);

/**
 * @brief Processes a video file and outputs roll corrections.
 */
void process_video(const std::string& path, int interval_ms, double area_percent);

} // namespace RollCorrection

#endif // FEATURE_BASED_HPP
