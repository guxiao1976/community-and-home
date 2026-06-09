#!/bin/bash
set -e
cd /home/jiaoxh/my-project/community-home

find services/ai-model-service/api/internal/handler -name "*.go" ! -name "*Zone*" ! -name "routes.go" -type f | while read f; do
  # Replace httpx.ErrorCtx(r.Context(), w, err) with response.Error(w, err)
  sed -i 's/httpx\.ErrorCtx(r\.Context(), w, err)/response.Error(w, err)/g' "$f"

  # Replace httpx.OkJsonCtx(r.Context(), w, resp) with response.Success(w, resp)
  sed -i 's/httpx\.OkJsonCtx(r\.Context(), w, resp)/response.Success(w, resp)/g' "$f"

  # Add response import after the svc import line
  sed -i '/services\/ai-model\/api\/internal\/svc/a\'$'\t''"github.com/guxiao/community-and-home/services/ai-model/api/internal/response"' "$f"

  echo "Updated: $f"
done
