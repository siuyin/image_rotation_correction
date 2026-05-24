(ns roll-corr.phase-correlation.core
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

(defn -main [& args]
  (let [video-path (or (first args) (str (System/getProperty "user.home") "/tennis1.mp4"))
        sampling-m (Integer/parseInt (or (second args) "1000"))
        p-area (Integer/parseInt (or (nth args 2 "75")))
        file (File. video-path)]
    (if (.exists file)
      (with-open [grab (FFmpegFrameGrabber. file)]
        (let [converter (OpenCVFrameConverter$ToMat.)]
          (.start grab)
          (let [first-frame (.grabImage grab)
                R (.convert converter first-frame)
                R-gray (Mat.)]
            (opencv_imgproc/cvtColor R R-gray opencv_imgproc/COLOR_BGR2GRAY)
            (let [R-central (get-central-area R-gray p-area)]
              (println "Starting Real-Time Phase Correlation loop on:" video-path)
              (loop [current-time-us 0]
                (.setTimestamp grab current-time-us)
                (if-let [frame (.grabImage grab)]
                  (let [V-n (.convert converter frame)
                        V-n-gray (Mat.)]
                    (opencv_imgproc/cvtColor V-n V-n-gray opencv_imgproc/COLOR_BGR2GRAY)
                    (let [V-n-central (get-central-area V-n-gray p-area)
                          _ (estimate-pitch-yaw R-central V-n-central)
                          {:keys [roll]} (estimate-zoom-roll R-central V-n-central)]
                      (printf "Time %.2fs - " (/ current-time-us 1000000.0))
                      (generate-report roll)
                      (recur (+ current-time-us (* sampling-m 1000)))))
                  (println "Reached end of video.")))))))
      (println "Video file not found:" video-path))))
