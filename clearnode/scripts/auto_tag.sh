#!/bin/sh

# Get the first arg to the script as version prefix
version_prefix=$1
if [ -z "$version_prefix" ]; then
  echo "Usage: $0 <version-prefix>"
  exit 1
fi

last_rc_tag=$(git ls-remote -t | grep -Eo "${version_prefix}-rc\.[0-9]+$" | sort -V -r | head -n 1)

new_rc_tag=""
if [ -z "$last_rc_tag" ]; then
  # No RC tags exist
  new_rc_tag="${version_prefix}-rc.0"
else
  # Last RC tag is the latest stable tag
  new_rc_tag="${version_prefix}-rc.$((${last_rc_tag##*.}+1))"
fi

git tag "$new_rc_tag"
git push origin "$new_rc_tag"
