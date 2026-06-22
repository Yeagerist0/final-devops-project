#!/usr/bin/env bash
set -euo pipefail
video_dir="$(cd "$(dirname "$0")" && pwd)"
for index in 0 1 2 3 4 5 6 7; do
  chromium --headless --no-sandbox --disable-gpu --disable-dev-shm-usage --hide-scrollbars \
    --window-size=1280,720 --screenshot="$video_dir/slide-$index.png" \
    "file://$video_dir/slides.html?slide=$index"
done
ffmpeg -y -loglevel error -framerate 1/10 -i "$video_dir/slide-%d.png" \
  -vf 'fps=24,format=yuv420p' -c:v libx264 -preset veryfast -crf 22 \
  -movflags +faststart "$video_dir/session21-demo.mp4"
