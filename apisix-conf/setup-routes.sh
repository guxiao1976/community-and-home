#!/bin/bash
# APISIX Route Setup Script
# Run this after APISIX container is started to configure routing via etcd.
# The routes are persisted in etcd, so this only needs to run once after etcd reset.
#
# Prerequisites:
#   - docker compose up -d (all middleware running)
#   - Backend services running on host.docker.internal ports

ETCD="http://127.0.0.1:2379"

etcd_put() {
  local key="$1"; local value="$2"
  local b64_key=$(echo -n "$key" | base64 -w0)
  local b64_value=$(echo -n "$value" | base64 -w0)
  curl -s -X POST "$ETCD/v3/kv/put" \
    -d "{\"key\":\"$b64_key\",\"value\":\"$b64_value\"}" > /dev/null
  echo "  OK: $key"
}

echo "=== Creating upstreams ==="

etcd_put "/apisix/upstreams/identity" \
  '{"name":"identity","type":"roundrobin","nodes":{"host.docker.internal:8888":1},"timeout":{"connect":3,"send":30,"read":30},"id":"identity"}'

etcd_put "/apisix/upstreams/masterdata" \
  '{"name":"masterdata","type":"roundrobin","nodes":{"host.docker.internal:8889":1},"timeout":{"connect":3,"send":30,"read":30},"id":"masterdata"}'

etcd_put "/apisix/upstreams/moderation" \
  '{"name":"moderation","type":"roundrobin","nodes":{"host.docker.internal:8890":1},"timeout":{"connect":3,"send":30,"read":30},"id":"moderation"}'

etcd_put "/apisix/upstreams/aimodel" \
  '{"name":"aimodel","type":"roundrobin","nodes":{"host.docker.internal:8891":1},"timeout":{"connect":3,"send":60,"read":60},"id":"aimodel"}'

etcd_put "/apisix/upstreams/auth" \
  '{"name":"auth","type":"roundrobin","nodes":{"host.docker.internal:8881":1},"timeout":{"connect":3,"send":30,"read":30},"id":"auth"}'

echo "=== Creating routes ==="

# Public — no JWT
etcd_put "/apisix/routes/identity-auth" \
  '{"uri":"/api/identity/auth/*","methods":["POST"],"upstream_id":"identity","priority":10,"id":"identity-auth","status":1}'

# Protected — JWT required
etcd_put "/apisix/routes/identity-protected" \
  '{"uri":"/api/identity/*","upstream_id":"identity","priority":5,"id":"identity-protected","status":1,"plugins":{"jwt-auth":{}}}'

etcd_put "/apisix/routes/masterdata" \
  '{"uri":"/api/masterdata/*","upstream_id":"masterdata","priority":5,"id":"masterdata","status":1,"plugins":{"jwt-auth":{}}}'

etcd_put "/apisix/routes/moderation" \
  '{"uri":"/api/moderation/*","upstream_id":"moderation","priority":5,"id":"moderation","status":1,"plugins":{"jwt-auth":{}}}'

etcd_put "/apisix/routes/aimodel" \
  '{"uri":"/api/v1/*","upstream_id":"aimodel","priority":5,"id":"aimodel","status":1,"plugins":{"jwt-auth":{}}}'

# Auth — REST API（不在 APISIX 层加 JWT；auth service 自行处理认证）
etcd_put "/apisix/routes/auth-api" \
  '{"uri":"/api/auth/*","methods":["GET","POST","PUT","DELETE"],"upstream_id":"auth","priority":10,"id":"auth-api","status":1}'

echo "=== Creating JWT consumer ==="

etcd_put "/apisix/consumers/community-user" \
  '{"username":"community-user","plugins":{"jwt-auth":{"key":"community-jwt","secret":"QBViQKNdUpsAdq48ClBoFxRoRwayo7MZFrLr_eu5NL8","algorithm":"HS256"}}}'

echo ""
echo "=== Done! Restart APISIX to apply: docker compose restart apisix ==="
echo ""
echo "To generate a test token:"
echo "  python3 -c \"import jwt,time; print(jwt.encode({'key':'community-jwt','user_id':1,'exp':int(time.time())+86400}, 'QBViQKNdUpsAdq48ClBoFxRoRwayo7MZFrLr_eu5NL8', algorithm='HS256'))\""
