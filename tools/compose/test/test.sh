#!/bin/sh
# shellcheck disable=SC2046,SC2119,SC2120

# exit on first error
set -e

. ./utils.sh

TENANT_ID=${1:-$(new_uuid)}
DEVICE_ID=${2:-$(new_uuid)}
FILE_LIST_URL="$FS_SERVER/api/internal/v1/files?tenant_id=$TENANT_ID&device_id=$DEVICE_ID"
TEST_FILE_COUNT=100

DEVICE_TOKEN_FILE=/tmp/device_token

# make device_token for api access
get_device_token() {
  curl -s "$DEVICE_TOKEN_URL?tenant_id=$TENANT_ID&device_id=$DEVICE_ID" >"$DEVICE_TOKEN_FILE"
  cat "$DEVICE_TOKEN_FILE"
}

# get upload urls
get_upload_urls() {
  count=${1:-$TEST_FILE_COUNT}
  echo "hello" >/tmp/1
  krypton-cli fs get_upload_url -server "$FS_SERVER" -jwt_token $(cat "$DEVICE_TOKEN_FILE") -count "$count"
  rm /tmp/1
}

# create files
create_files() {
  count=${1:-$TEST_FILE_COUNT}
  krypton-cli fs create_file -server "$FS_SERVER" -jwt_token $(cat "$DEVICE_TOKEN_FILE") -count "$count"
}

# download_url test
download_url() {
  id=${1:-101}
  krypton-cli fs get_download_url -server "$FS_SERVER" -file_id "$id"
}

# fetch details of a known file
get_file_details() {
  id=${1:-101}
  krypton-cli fs get_file_details -server "$FS_SERVER" -file_id "$id" -jwt_token $(cat "$DEVICE_TOKEN_FILE")
}

wait_for_server
get_device_token
get_upload_urls

# get number of files marked uploaded before create tests
# this is used as a baseline to check newly created files
CURRENT_UPLOADED_COUNT=$(count_uploaded_files "$FILE_LIST_URL" 5)
echo "Current uploaded files: $CURRENT_UPLOADED_COUNT"

create_files
#check_uploaded_count "$FILE_LIST_URL" $((CURRENT_UPLOADED_COUNT + TEST_FILE_COUNT))
download_url
#get_file_details
