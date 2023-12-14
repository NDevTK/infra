#!/usr/bin/env bash

set -x
set -e

DEPOT_TOOLS_URL=https://chromium.googlesource.com/chromium/tools/depot_tools

# CD to script dir.
cd "${0%/*}"

depot_tools=`mktemp -d`
function clean {
  rm -rf "$depot_tools"
}
trap clean EXIT

git -C "$depot_tools" init
git -C "$depot_tools" pull --depth 1 "$DEPOT_TOOLS_URL"

mkdir -p third_party

sed 's/import gclient_utils//g' "$depot_tools/gclient_eval.py" > "$depot_tools/gclient_eval.py.2"
sed 's/gclient_utils.Error/ValueError/g' "$depot_tools/gclient_eval.py.2" > "$depot_tools/gclient_eval.py.3"

cp "$depot_tools/gclient_eval.py.3" gclient_eval.py
cp -a "$depot_tools/third_party/schema" ./third_party

echo "Source URL: $DEPOT_TOOLS_URL" > update.record
echo "Source Commit: $(git -C "$depot_tools" rev-parse FETCH_HEAD)" >> update.record
echo "Edits:" >> update.record
echo "  * Remove gclient_utils import from gclient_eval.py" >> update.record
echo "  * Replace gclient_utils.Error in gclient_eval with ValueError" >> update.record
