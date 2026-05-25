#include "feature_based.hpp"
#include <iostream>
#include <string>

void print_usage() {
    std::cout << "Usage: roll_corr VIDEO_PATH [-i INTERVAL_MS] [-a AREA_PERCENT]" << std::endl;
}

bool try_parse_interval(const std::string& arg, int& argc, int& i, char** argv, int& interval) {
    if (arg != "-i" || i + 1 >= argc) return false;
    interval = std::stoi(argv[++i]);
    return true;
}

bool try_parse_area(const std::string& arg, int& argc, int& i, char** argv, double& area) {
    if (arg != "-a" || i + 1 >= argc) return false;
    area = std::stod(argv[++i]);
    return true;
}

int main(int argc, char** argv) {
    if (argc < 2) {
        print_usage();
        return 1;
    }

    std::string video_path = argv[1];
    int interval = 1000;
    double area = 75.0;

    for (int i = 2; i < argc; ++i) {
        std::string arg = argv[i];
        
        if (try_parse_interval(arg, argc, i, argv, interval)) continue;
        if (try_parse_area(arg, argc, i, argv, area)) continue;

        print_usage();
        return 1;
    }

    try {
        RollCorrection::process_video(video_path, interval, area);
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }

    return 0;
}
