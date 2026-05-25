cmake_minimum_required(VERSION 3.10)
project(roll_corr_feature_based)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find OpenCV
find_package(OpenCV REQUIRED)
include_directories(${OpenCV_INCLUDE_DIRS})
include_directories(include)

# Core library
add_library(feature_based src/feature_based.cpp)
target_link_libraries(feature_based ${OpenCV_LIBS})

# Executable
add_executable(roll_corr src/main.cpp)
target_link_libraries(roll_corr feature_based ${OpenCV_LIBS})

# Tests
enable_testing()
add_executable(test_feature_based test/test_feature_based.cpp)
target_link_libraries(test_feature_based feature_based ${OpenCV_LIBS})
add_test(NAME FeatureBasedTests COMMAND test_feature_based)
