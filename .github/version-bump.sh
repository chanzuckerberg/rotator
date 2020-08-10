#!/bin/bash -ex
git config --global user.email "no-reply@chanzuckerberg.com"
git config --global user.name "GitHub Actions Bot"

awk -i inplace 'BEGIN { FS = "." } ; { print $1 "." $2 "." ++$3 }' VERSION
version=$(cat VERSION)
sed -i "s/appVersion:.*/appVersion: v${version}/" charts/Chart.yaml

git add VERSION charts/Chart.yaml
git commit -m "release version ${version}"
git tag v"${version}"

# NOTE - This push can fail if someone pushed to master while the build
#  was running. We're choosing to mostly ignore this situation due to our
#  currently fairly low commit velocity.
commit_hash=$(git rev-parse --short HEAD)
git push origin ${commit_hash}:master --tags
