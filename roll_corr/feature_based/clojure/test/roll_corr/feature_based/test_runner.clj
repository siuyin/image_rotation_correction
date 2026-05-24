(ns roll-corr.feature-based.test-runner
  (:require [clojure.test :as t]
            [clojure.java.io :as io]
            [clojure.string :as str]))

(defn- find-test-namespaces []
  (->> (file-seq (io/file "test"))
       (filter #(and (.isFile %) (str/ends-with? (.getName %) "_test.clj")))
       (map #(-> (.getPath %)
                 (str/replace #"^test/" "")
                 (str/replace #"_test\.clj$" "_test")
                 (str/replace #"/" ".")
                 (str/replace #"_" "-")))
       (map symbol)))

(defn run [_]
  (let [nss (find-test-namespaces)]
    (apply require nss)
    (let [summary (apply t/run-tests nss)]
      (when (pos? (+ (:fail summary) (:error summary)))
        (System/exit 1)))))
