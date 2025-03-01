#!/usr/bin/env bash
set -e

# 1. Clone or download the repo (or skip if user already did)
#    For demonstration, we'll clone into a temp dir:
TMP_DIR=$(mktemp -d)
git clone --depth=1 git@github.com:joeeeeey/ai-commit.git "$TMP_DIR"

# 2. Build the binary (or download one from releases)
cd "$TMP_DIR"
make build

# 3. Copy the hook scripts & binary into the user's .git/hooks/ of the current repo
#    Make sure you're in a valid git repository
if [ ! -d .git ]; then
  echo "Looks like you're not in a Git repository. Please run this in your target repo."
  exit 1
fi

make create_hook

echo "ai-commit installed! Remember to set your AI_COMMIT_TOKEN environment variable."
echo "e.g. export AI_COMMIT_TOKEN='YOUR_KEY_HERE'"