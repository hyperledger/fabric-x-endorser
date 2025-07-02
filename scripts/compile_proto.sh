#!/usr/bin/env bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -euo pipefail

# Find all proto dirs to be processed
PROTO_DIRS="$(find . \
    -path ./protos -prune -o \
    -path ./wgclient -prune -o \
    -name '*.proto' -print0 | \
    xargs -0 -n 1 dirname | \
    sort -u | grep -v testdata)"

for dir in ${PROTO_DIRS}; do
  echo "Compiling: $dir"
  protoc --proto_path="${PWD}" \
      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
      --go_out=paths=source_relative:. "${PWD}/$dir"/*.proto
done
