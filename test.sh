#! /bin/bash

curl -X POST 'https://api.dify.ai/v1/workflows/run' \
--header "Authorization: Bearer ${api_key}" \
--header 'Content-Type: application/json' \
--data-raw '{
    "inputs": {
      "repo_name": "k8s",
      "diff_text": "-    local prodReplicas = if context.Region == 'us-west-1' then 9 else 1, \n+    local prodReplicas = if context.Region == 'us-west-1' then 1 else 1,"
  },
    "response_mode": "blocking",
    "user": "abc-123"
}'




    # data = {
    #     "inputs": {
    #         "orig_mail": {
    #             "transfer_method": "local_file",
    #             "upload_file_id": file_id,
    #             "type": "document"
    #         }
    #     },
    #     "response_mode": response_mode,
    #     "user": user
    # }