#!/bin/bash

set -e

bump_type="${1:-patch}"

case "$bump_type" in
  patch|minor|major) ;;
  *)
    echo "Error: Invalid bump type '$bump_type'. Use 'patch', 'minor', or 'major'."
    exit 1
    ;;
esac

# Collect all tags from local and remote (strip annotated tag peel refs like ^{})
all_tags=$( (git tag; git ls-remote --tags origin) 2>/dev/null | sed 's|.*refs/tags/||; s|\^{}||' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | uniq )

if [ -z "$all_tags" ]; then
  echo "Error: No semver tags (vX.Y.Z) found locally or on origin."
  exit 1
fi

latest_tag=$(echo "$all_tags" | tail -1)
latest_ver=${latest_tag#v}
IFS='.' read -r major minor patch <<< "$latest_ver"

case "$bump_type" in
  patch)
    patch=$((patch + 1))
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
esac

new_tag="v${major}.${minor}.${patch}"

# Check if the tag already exists locally
if git tag --list | grep -q "^$new_tag$"; then
  echo "Error: Tag '$new_tag' already exists locally."
  exit 1
fi

# Check if the tag already exists remotely
if git ls-remote --tags origin | grep -q "refs/tags/$new_tag$"; then
  echo "Error: Tag '$new_tag' already exists remotely."
  exit 1
fi

# Create the tag with a message
if ! git tag -a "$new_tag" -m "Release $new_tag"; then
  echo "Error: Failed to create tag '$new_tag'."
  exit 1
fi

# Push the tag to the remote repository
if ! git push origin "$new_tag"; then
  echo "Error: Failed to push tag '$new_tag' to origin."
  exit 1
fi

echo "Tag '$new_tag' created (bumped from '$latest_tag' by $bump_type) and pushed successfully."
exit 0
