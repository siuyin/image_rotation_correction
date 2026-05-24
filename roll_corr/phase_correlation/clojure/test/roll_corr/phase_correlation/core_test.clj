(ns roll-corr.phase-correlation.core-test
  (:require [clojure.test :refer [deftest is testing]]
            [roll-corr.phase-correlation.core :as core]))

;; Fundamentals approach: Representing an image as a 2D vector of numbers (intensities)
(def sample-img
  [[0 0 0 0]
   [0 1 1 0]
   [0 1 1 0]
   [0 0 0 0]])

(deftest get-central-area-test
  (testing "Extraction of 25% area from 4x4 image (results in 2x2)"
    (let [central (core/get-central-area sample-img 25)]
      (is (= 2 (count central)))
      (is (= 2 (count (first central))))
      (is (= [[1 1] [1 1]] central)))))

(deftest to-log-polar-test
  (testing "Stub for Log-Polar transformation"
    (is (vector? (core/to-log-polar sample-img)))))

(deftest compute-phase-correlation-test
  (testing "Zero shift between identical images"
    (is (= [0.0 0.0] (core/compute-phase-correlation sample-img sample-img)))))

(deftest estimate-pitch-yaw-test
  (testing "Estimation of tilt and pan"
    (let [result (core/estimate-pitch-yaw sample-img sample-img)]
      (is (contains? result :pitch))
      (is (contains? result :yaw)))))

(deftest estimate-zoom-roll-test
  (testing "Estimation of zoom and roll"
    (let [result (core/estimate-zoom-roll sample-img sample-img)]
      (is (contains? result :zoom))
      (is (contains? result :roll)))))
