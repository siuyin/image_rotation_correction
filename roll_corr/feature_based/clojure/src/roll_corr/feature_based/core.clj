(ns roll-corr.feature-based.core
  (:gen-class)
  (:require [clojure.string :as str])
  (:import [java.io File]
           [org.bytedeco.javacv FFmpegFrameGrabber OpenCVFrameConverter$ToMat]
           [org.bytedeco.opencv.opencv_core Mat KeyPoint DMatch Rect Point2f Point2fVector KeyPointVector DMatchVector]
           [org.bytedeco.opencv.opencv_features2d ORB DescriptorMatcher]
           [org.bytedeco.opencv.global opencv_core opencv_imgproc opencv_features2d opencv_calib3d]))

(defn get-central-area [^Mat image p-area]
  (let [h (.rows image)
        w (.cols image)
        ratio (Math/sqrt (/ p-area 100.0))
        new-w (int (* w ratio))
        new-h (int (* h ratio))
        x (quot (- w new-w) 2)
        y (quot (- h new-h) 2)]
    (Mat. image (Rect. x y new-w new-h))))

(defn detect-and-describe [^Mat image]
  (let [orb (ORB/create)
        keypoints (KeyPointVector.)
        descriptors (Mat.)]
    (.detectAndCompute orb image (Mat.) keypoints descriptors)
    [keypoints descriptors]))

(defn match-features [descriptors-R descriptors-V]
  (let [matcher (DescriptorMatcher/create DescriptorMatcher/BRUTEFORCE_HAMMING)
        matches (DMatchVector.)]
    (.match matcher descriptors-R descriptors-V matches)
    matches))

(defn solve-geometric-mapping [kp-R kp-V matches]
  (let [n (.size matches)]
    (if (>= n 4)
      (let [src-mat (Mat. n 2 opencv_core/CV_32F)
            dst-mat (Mat. n 2 opencv_core/CV_32F)
            src-indexer (.createIndexer src-mat)
            dst-indexer (.createIndexer dst-mat)
            _ (doseq [i (range n)]
                (let [m (.get matches i)
                      ptR (.pt (.get ^KeyPointVector kp-R (.queryIdx m)))
                      ptV (.pt (.get ^KeyPointVector kp-V (.trainIdx m)))]
                  (.put src-indexer i 0 (float (.x ptR)))
                  (.put src-indexer i 1 (float (.y ptR)))
                  (.put dst-indexer i 0 (float (.x ptV)))
                  (.put dst-indexer i 1 (float (.y ptV)))))
            mapping (opencv_calib3d/findHomography src-mat dst-mat (Mat.) opencv_calib3d/RANSAC 3.0)]
        mapping)
      nil)))

(defn decompose-transformation [^Mat mapping]
  (if (and mapping (not (.empty mapping)))
    (let [indexer (.createIndexer mapping)
          m00 (.get indexer (long-array [0 0]))
          m10 (.get indexer (long-array [1 0]))]
      {:pitch 0.0 :yaw 0.0 :zoom 1.0 
       :roll (Math/toDegrees (Math/atan2 m10 m00))})
    {:pitch 0.0 :yaw 0.0 :zoom 1.0 :roll 0.0}))

(defn generate-report [roll-angle]
  (println (format "Required Roll Correction: %.2f degrees" roll-angle)))

(defn- print-usage []
  (println "Usage: clojure -M -m roll-corr.feature-based.core VIDEO_PATH [flags]")
  (println "  VIDEO_PATH            Path to video file (REQUIRED)")
  (println "  -i, --interval MS     Sampling interval in ms (default: 1000)")
  (println "  -a, --area PERCENT    Central area percentage (default: 75)")
  (println "  -h, --help            Print this help"))

(defn- get-env-or-default [key default]
  (let [val (System/getenv key)]
    (if val (Integer/parseInt val) default)))

(defn- parse-args [args]
  (let [path (first (remove #(str/starts-with? % "-") args))
        options (loop [args (remove #(= % path) args)
                       options {:path path
                                :interval (get-env-or-default "INTERVAL" 1000)
                                :area (get-env-or-default "AREA" 75)}]
                  (cond
                    (empty? args) options
                    (#{ "-h" "--help" } (first args)) (assoc options :help true)
                    (#{ "-i" "--interval" } (first args)) (recur (drop 2 args) (assoc options :interval (Integer/parseInt (second args))))
                    (#{ "-a" "--area" } (first args)) (recur (drop 2 args) (assoc options :area (Integer/parseInt (second args))))
                    :else options))]
    options))

(defn- open-video [path]
  (let [grabber (FFmpegFrameGrabber. path)]
    (.start grabber)
    grabber))

(defn- grab-reference [grabber area]
  (let [converter (OpenCVFrameConverter$ToMat.)
        frame (.grabImage grabber)
        img (.convert converter frame)
        gray (Mat.)]
    (opencv_imgproc/cvtColor img gray opencv_imgproc/COLOR_BGR2GRAY)
    (get-central-area gray area)))

(defn- process-stream [grabber ref-central ref-kp ref-desc interval area]
  (let [converter (OpenCVFrameConverter$ToMat.)]
    (loop [current-time-us 0]
      (.setTimestamp grabber current-time-us)
      (if-let [frame (.grabImage grabber)]
        (let [img (.convert converter frame)
              gray (Mat.)]
          (opencv_imgproc/cvtColor img gray opencv_imgproc/COLOR_BGR2GRAY)
          (let [central (get-central-area gray area)
                [kp-V desc-V] (detect-and-describe central)
                matches (match-features ref-desc desc-V)
                mapping (solve-geometric-mapping ref-kp kp-V matches)
                {:keys [roll]} (decompose-transformation mapping)]
            (printf "Time %.2fs - " (/ current-time-us 1000000.0))
            (generate-report roll)
            (recur (+ current-time-us (* interval 1000)))))
        (println "Reached end of video.")))))

(defn -main [& args]
  (let [{:keys [path interval area help]} (parse-args args)]
    (if (or help (not path) (not (.exists (File. path))))
      (print-usage)
      (with-open [grabber (doto (FFmpegFrameGrabber. path)
                            (.setOption "loglevel" "quiet"))]
        (.start grabber)
        (let [ref (grab-reference grabber area)
              [ref-kp ref-desc] (detect-and-describe ref)]
          (process-stream grabber ref ref-kp ref-desc interval area))))))
