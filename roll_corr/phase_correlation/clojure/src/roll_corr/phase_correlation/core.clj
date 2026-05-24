(ns roll-corr.phase-correlation.core
  (:gen-class)
  (:require [clojure.string :as str])
  (:import [java.io File]
           [org.bytedeco.javacv FFmpegFrameGrabber OpenCVFrameConverter$ToMat]
           [org.bytedeco.opencv.opencv_core Mat Point2f Size Rect]
           [org.bytedeco.opencv.global opencv_core opencv_imgproc]))

(defn get-central-area [^Mat image p-area]
  (let [h (.rows image)
        w (.cols image)
        ratio (Math/sqrt (/ p-area 100.0))
        new-w (int (* w ratio))
        new-h (int (* h ratio))
        x (quot (- w new-w) 2)
        y (quot (- h new-h) 2)]
    (Mat. image (Rect. x y new-w new-h))))

(defn to-log-polar [^Mat image]
  (let [h (.rows image)
        w (.cols image)
        center (Point2f. (/ w 2.0) (/ h 2.0))
        max-radius (double (/ (Math/sqrt (+ (* h h) (* w w))) 2.0))
        dest (Mat.)
        flags (int (bit-or opencv_imgproc/INTER_LINEAR opencv_imgproc/WARP_POLAR_LOG))]
    (opencv_imgproc/warpPolar image dest (Size. w h) center max-radius flags)
    dest))

(defn compute-phase-correlation [^Mat img1 ^Mat img2]
  (let [img1-32f (Mat.)
        img2-32f (Mat.)]
    (.convertTo img1 img1-32f opencv_core/CV_32F)
    (.convertTo img2 img2-32f opencv_core/CV_32F)
    (let [point (opencv_imgproc/phaseCorrelate img1-32f img2-32f)]
      [(.x point) (.y point)])))

(defn estimate-pitch-yaw [^Mat R-central ^Mat V-n-central]
  (let [[dx dy] (compute-phase-correlation R-central V-n-central)]
    {:pitch dy :yaw dx}))

(defn estimate-zoom-roll [^Mat R-central ^Mat V-n-central]
  (let [lp-R (to-log-polar R-central)
        lp-V (to-log-polar V-n-central)
        [d-log-r d-theta] (compute-phase-correlation lp-R lp-V)
        h (.rows R-central)
        w (.cols R-central)
        max-radius (/ (Math/sqrt (+ (* h h) (* w w))) 2.0)
        ;; Map d-theta back to degrees
        roll-deg (* d-theta (/ 360.0 h))]
    {:zoom (Math/exp (* d-log-r (/ (Math/log max-radius) w))) 
     :roll roll-deg}))

(defn generate-report [roll-angle]
  (println (format "Required Roll Correction: %.2f degrees" roll-angle)))
(defn- print-usage []
  (println "Usage: clojure -M -m roll-corr.phase-correlation.core VIDEO_PATH [flags]")
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

(defn- process-stream [grabber ref-central interval area]
  (let [converter (OpenCVFrameConverter$ToMat.)]
    (loop [current-time-us 0]
      (.setTimestamp grabber current-time-us)
      (if-let [frame (.grabImage grabber)]
        (let [img (.convert converter frame)
              gray (Mat.)]
          (opencv_imgproc/cvtColor img gray opencv_imgproc/COLOR_BGR2GRAY)
          (let [central (get-central-area gray area)
                _ (estimate-pitch-yaw ref-central central)
                {:keys [roll]} (estimate-zoom-roll ref-central central)]
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
        (let [ref (grab-reference grabber area)]
          (process-stream grabber ref interval area))))))
