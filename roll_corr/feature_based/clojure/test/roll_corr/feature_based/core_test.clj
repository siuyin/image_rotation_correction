(ns roll-corr.feature-based.core-test
  (:require [clojure.test :refer [deftest is testing]]
            [roll-corr.feature-based.core :as core]))

(def sample-img
  [[0 0 0 0]
   [0 1 1 0]
   [0 1 1 0]
   [0 0 0 0]])

(deftest get-central-area-test
  (testing "Extraction of 25% area from 4x4 image"
    (let [central (core/get-central-area sample-img 25)]
      (is (= 2 (count central)))
      (is (= 2 (count (first central)))))))

(deftest detect-and-describe-test
  (testing "Detecting key points and signatures"
    (let [results (core/detect-and-describe sample-img)]
      (is (sequential? results))
      (if (seq results)
        (let [first-pt (first results)]
          (is (contains? first-pt :point))
          (is (contains? first-pt :signature)))))))

(deftest match-features-test
  (testing "Matching signatures between images"
    (let [sigs [{:point [1 1] :signature [1 0 1]}]
          matches (core/match-features sigs sigs)]
      (is (= 1 (count matches)))
      (is (= [[1 1] [1 1]] (first matches))))))

(deftest solve-geometric-mapping-test
  (testing "Solving for transformation matrix"
    (let [matches [[[0 0] [0 0]] [[1 1] [1 1]] [[2 2] [2 2]] [[3 3] [3 3]]]
          matrix (core/solve-geometric-mapping matches)]
      (is (= 3 (count matrix)))
      (is (= 3 (count (first matrix)))))))

(deftest decompose-transformation-test
  (testing "Decomposing mapping into camera parameters"
    (let [matrix [[1 0 0] [0 1 0] [0 0 1]]
          params (core/decompose-transformation matrix)]
      (is (contains? params :pitch))
      (is (contains? params :yaw))
      (is (contains? params :zoom))
      (is (contains? params :roll)))))
