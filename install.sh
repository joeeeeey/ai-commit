#!/usr/bin/env bash
set -e

# --------------- CONFIG ---------------
REPO_OWNER="joeeeeey"
REPO_NAME="ai-commit"
RELEASE_TAG="v0.0.1"  # Update with your release tag name

# Construct the base URL for Releases
# https://github.com/joeeeeey/ai-commit/releases/download/v0.0.1
BASE_DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${RELEASE_TAG}"

# The path to the hook script in your repo:
# https://raw.githubusercontent.com/joeeeeey/ai-commit/refs/heads/main/hook/prepare-commit-msg
HOOK_URL="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/refs/heads/main/hook/prepare-commit-msg"

# --------------- FUNCTIONS ---------------

detect_os() {
  uname_str="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$uname_str" in
    linux*)   os="linux" ;;
    darwin*)  os="darwin" ;;
    msys*|cygwin*|mingw*) os="windows" ;;
    *)        os="unsupported" ;;
  esac
  echo "$os"
}

detect_arch() {
  arch_str="$(uname -m)"
  case "$arch_str" in
    x86_64)  arch="amd64" ;;
    arm64)   arch="arm64" ;;
    *)       arch="unsupported" ;;
  esac
  echo "$arch"
}

# --------------- MAIN SCRIPT ---------------

# 1) Check if .git folder exists
if [ ! -d ".git" ]; then
  echo "Error: .git directory not found. Please run this within a Git repository."
  exit 1
fi

# 2) Detect OS/Arch
OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
  echo "Error: unsupported OS/arch. Please check the GitHub releases for a compatible binary."
  exit 1
fi

# 3) Determine correct binary name
FILE_NAME="commit_msg_generator_${OS}_${ARCH}"
if [ "$OS" = "windows" ]; then
  FILE_NAME="commit_msg_generator_win_amd64.exe"
fi

BINARY_URL="${BASE_DOWNLOAD_URL}/${FILE_NAME}"
# https://github.com/joeeeeey/ai-commit/releases/download/v0.0.1/commit_msg_generator_darwin_arm64

echo "Detected OS=$OS, ARCH=$ARCH"
echo "Downloading from $BINARY_URL ..."

# 4) Download the binary into .git/hooks/
curl -L --fail -o ".git/hooks/commit_msg_generator" "$BINARY_URL"
chmod +x ".git/hooks/commit_msg_generator"

echo "Binary downloaded and made executable."

# 5) Download the prepare-commit-msg script
echo "Downloading hook script from $HOOK_URL ..."
curl -L --fail -o ".git/hooks/prepare-commit-msg" "$HOOK_URL"
chmod +x ".git/hooks/prepare-commit-msg"

echo "Ai-commit Hook installed! Remember to set your AI_COMMIT_TOKEN environment variable."
echo "e.g. export AI_COMMIT_TOKEN='YOUR_KEY_HERE' (Or add this to your ~/.bashrc, ~/.zshrc, etc.)"