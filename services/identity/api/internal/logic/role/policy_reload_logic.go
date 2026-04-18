package role

import (
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	PolicyReloadChannel = "casbin:policy:reload"
)

type PolicyReloadMessage struct {
	RoleId int64  `json:"role_id"`
	Action string `json:"action"` // "update", "delete"
}

// PublishPolicyReload publishes a policy reload message to Redis
func PublishPolicyReload(rds *redis.Redis, roleId int64, action string) error {
	msg := PolicyReloadMessage{
		RoleId: roleId,
		Action: action,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logx.Errorf("Failed to marshal policy reload message: %v", err)
		return err
	}

	_, err = rds.Publish(PolicyReloadChannel, string(data))
	if err != nil {
		logx.Errorf("Failed to publish policy reload message: %v", err)
		return err
	}

	logx.Infof("Published policy reload message for role %d, action: %s", roleId, action)
	return nil
}

// TODO: Implement SubscribePolicyReload using go-redis client
// The go-zero Redis wrapper doesn't support Subscribe/PubSub operations
// Need to use github.com/redis/go-redis/v9 for subscription functionality
