#include "feature_based.hpp"
#include <iostream>
#include <cassert>
#include <cmath>

void test_get_central_area() {
    std::cout << "Testing get_central_area..." << std::endl;
    cv::Mat img = cv::Mat::zeros(100, 100, CV_8UC1);
    cv::Mat central = RollCorrection::get_central_area(img, 25.0); // 25% area -> 50x50
    assert(central.rows == 50);
    assert(central.cols == 50);
    std::cout << "  Passed!" << std::endl;
}

void test_detect_and_describe() {
    std::cout << "Testing detect_and_describe..." << std::endl;
    cv::Mat img = cv::Mat::zeros(100, 100, CV_8UC1);
    cv::rectangle(img, cv::Rect(40, 40, 20, 20), cv::Scalar(255), -1);
    
    std::vector<cv::KeyPoint> kp;
    cv::Mat desc;
    RollCorrection::detect_and_describe(img, kp, desc);
    
    // We expect some keypoints for a simple square
    assert(!kp.empty());
    assert(desc.rows == (int)kp.size());
    std::cout << "  Passed!" << std::endl;
}

void test_decompose_transformation() {
    std::cout << "Testing decompose_transformation..." << std::endl;
    
    // Identity mapping (0 degrees)
    cv::Mat identity = (cv::Mat_<double>(3,3) << 1, 0, 0, 0, 1, 0, 0, 0, 1);
    auto params = RollCorrection::decompose_transformation(identity);
    assert(std::abs(params.roll) < 1e-6);

    // 90 degree rotation
    // [ cos(90) -sin(90) 0 ]   [ 0 -1 0 ]
    // [ sin(90)  cos(90) 0 ] = [ 1  0 0 ]
    // [ 0        0       1 ]   [ 0  0 1 ]
    cv::Mat rot90 = (cv::Mat_<double>(3,3) << 0, -1, 0, 1, 0, 0, 0, 0, 1);
    params = RollCorrection::decompose_transformation(rot90);
    assert(std::abs(params.roll - 90.0) < 1e-6);

    // -45 degree rotation
    double angle = -45.0 * CV_PI / 180.0;
    double c = std::cos(angle);
    double s = std::sin(angle);
    cv::Mat rot45neg = (cv::Mat_<double>(3,3) << c, -s, 0, s, c, 0, 0, 0, 1);
    params = RollCorrection::decompose_transformation(rot45neg);
    assert(std::abs(params.roll - (-45.0)) < 1e-6);

    std::cout << "  Passed!" << std::endl;
}

int main() {
    test_get_central_area();
    test_detect_and_describe();
    test_decompose_transformation();
    std::cout << "All tests passed!" << std::endl;
    return 0;
}
