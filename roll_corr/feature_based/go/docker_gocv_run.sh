#!/bin/bash
docker run -it --name gocv -h gocv -v $HOME:/h -u $UID --network host \
  -v /tmp/.X11-unix:/tmp/.X11-unix \
  -e DISPLAY=$DISPLAY -w /h gocv/opencv bash