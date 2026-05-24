(ns roll-corr.feature-based.core
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
  (if (>= (.size matches) 4)
    (let [n (.size matches)
          src-mat (Mat. n 2 opencv_core/CV_32F)
          dst-mat (Mat. n 2 opencv_core/CV_32F)
          src-indexer (.createIndexer src-mat)
          dst-indexer (.createIndexer dst-mat)
          _ (doseq [i (range n)]
              (let [m (.get matches i)
                    ptR (.pt (.get kp-R (.queryIdx m)))
                    ptV (.pt (.get kp-V (.trainIdx m)))]
                (.put src-indexer i 0 (float (.x ptR)))
                (.put src-indexer i 1 (float (.y ptR)))
                (.put dst-indexer i 0 (float (.x ptV)))
                (.put dst-indexer i 1 (float (.y ptV)))))
          mapping (opencv_calib3d/findHomography src-mat dst-mat (Mat.) opencv_calib3d/RANSAC 3.0)]
      mapping)
    nil))

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
            (let [R-central (get-central-area R-gray p-area)
                  [kp-R desc-R] (detect-and-describe R-central)]
              (println "Starting Real-Time Feature-Based loop on:" video-path)
              (loop [current-time-us 0]
                (.setTimestamp grab current-time-us)
                (if-let [frame (.grabImage grab)]
                  (let [V-n (.convert converter frame)
                        V-n-gray (Mat.)]
                    (opencv_imgproc/cvtColor V-n V-n-gray opencv_imgproc/COLOR_BGR2GRAY)
                    (let [V-n-central (get-central-area V-n-gray p-area)
                          [kp-V desc-V] (detect-and-describe V-n-central)
                          matches (match-features desc-R desc-V)
                          mapping (solve-geometric-mapping kp-R kp-V matches)
                          {:keys [roll]} (decompose-transformation mapping)]
                      (printf "Time %.2fs - " (/ current-time-us 1000000.0))
                      (generate-report roll)
                      (recur (+ current-time-us (* sampling-m 1000)))))
                  (println "Reached end of video.")))))))
      (println "Video file not found:" video-path))))
