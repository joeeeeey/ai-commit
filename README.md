# ai-commit

**Automatically generate commit messages with AI** whenever you run `git commit`.  
This repo provides a cross-platform Go binary and a Git `prepare-commit-msg` hook that integrates an AI prompt.

---

## Demo

Here's a quick GIF showing how `ai-commit` works in action:

![demo gif](./static/demo.gif)  

---

## Quick Install

### Install by CURL
The following command will auto add git-hook under the current directory with .git

```bash
curl -sL https://raw.githubusercontent.com/joeeeeey/ai-commit/main/install.sh | bash
```

## Uninstall
```bash
# cd $dir_with_git
rm .git/hooks/prepare-commit-msg
rm .git/hooks/commit_msg_generator
```

## Test by curl
```bash

curl -X POST 'https://api.dify.ai/v1/workflows/run' \
--header "Authorization: Bearer ${AI_COMMIT_TOKEN}" \
--header 'Content-Type: application/json' \
--data-raw '{
    "inputs": {
      "repo_name": "k8s",
      "diff_text": "-    local prodReplicas = if context.Region == 'us-west-1' then 9 else 1, \n+    local prodReplicas = if context.Region == 'us-west-1' then 1 else 1,"
  },
    "response_mode": "blocking",
    "user": "abc-123"
}'
```