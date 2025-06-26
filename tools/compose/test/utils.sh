#!/bin/sh
# count all uploaded files with status as "U". U = uploaded.
# Files are marked uploaded when a file upload notification is processed
# via file store and a corresponding queue notification is processed
# by the file service. Appropriate wait times are needed as this
# is asynchronous

# count all uploaded files with status as "U".
count_uploaded_files() {
  url="$1"
  wait_duration=${2:-5}

  sleep "$wait_duration"
  curl -s "$url" | jq '.files[] | select(.status=="uploaded") | .status' | wc -l
}

wait_and_check_uploaded_count() {
  url="$1"
  expected_count="$2"
  wait_duration=${3:-3}
  retry_count=${4:-20}

  for i in $(seq 1 "$retry_count"); do
    uploaded_count=$(count_uploaded_files "$url" "$wait_duration")
    echo "Files marked as uploaded: $uploaded_count / $expected_count"
    if [ "$uploaded_count" -ne "$expected_count" ]; then
      echo "Wait $i/$retry_count: Expected: $expected_count. Got $uploaded_count."
      sleep "$wait_duration"
    else
      break
    fi
  done
  if [ "$uploaded_count" -ne "$expected_count" ]; then
    echo "Error: Expected: $expected_count. Got $uploaded_count."
    exit 1
  fi
}

wait_for_queue_stabilize() {
  url="$1"
  expected_count=${2:-0}
  wait_duration=${3:-5}
  uploaded_count=$(count_uploaded_files "$url" "$wait_duration")
  echo "Files marked as uploaded: $uploaded_count / $expected_count"
  if [ "$uploaded_count" -ne "$expected_count" ]; then
    echo "Error: Expected: $expected_count. Got $uploaded_count."
    exit 1
  fi
}

# count uploaded files and wait for queue to settle, then verify count
check_uploaded_count() {
  url="$1"
  expected_count="$2"
  wait_and_check_uploaded_count "$url" "$expected_count"
  wait_for_queue_stabilize "$url" "$expected_count"
}

new_uuid() {
  cat /proc/sys/kernel/random/uuid
}
