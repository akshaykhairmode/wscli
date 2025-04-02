#!/bin/bash

tag_name="$1"

if [ -z "$tag_name" ]; then
  echo "Error: Tag name is required."
  exit 1
fi

# Check if the tag already exists locally
if git tag --list | grep -q "^$tag_name$"; then
  echo "Error: Tag '$tag_name' already exists locally."
  exit 1
fi

# Check if the tag already exists remotely
if git ls-remote --tags origin | grep -q "refs/tags/$tag_name$"; then
  echo "Error: Tag '$tag_name' already exists remotely."
  exit 1
fi

# Create the tag with a message
if ! git tag -a "$tag_name" -m "Release $tag_name"; then
  echo "Error: Failed to create tag '$tag_name'."
  exit 1
fi

# Push the tag to the remote repository
if ! git push origin "$tag_name"; then
  echo "Error: Failed to push tag '$tag_name' to origin."
  exit 1
fi

echo "Tag '$tag_name' created and pushed successfully."
exit 0
