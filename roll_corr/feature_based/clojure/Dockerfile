# Stage 1: Build
FROM clojure:temurin-17-tools-deps-1.12.0.1479 AS build
WORKDIR /app
COPY . .
RUN cd roll_corr/feature_based/clojure && clojure -T:build uber

# Stage 2: Run
FROM eclipse-temurin:17-jre
RUN apt-get update && apt-get install -y libgtk2.0-0 libgomp1 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
# Copy only the standalone JAR from the build stage
COPY --from=build /app/roll_corr/feature_based/clojure/target/app-1.0.0-standalone.jar app.jar
# ENV variables allow runtime configuration
ENV INTERVAL=1000
ENV AREA=75
ENTRYPOINT ["java", "-jar", "app.jar"]
