// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"context"
	"time"

	"github.com/HPInc/krypton-fs/service/common"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// Name of the Redis key to be used as a leader lock.
	leaderLockName = "krypton-fs-leader"

	// TTL set for the leader lock. It should expire after this duration.
	LeaderLockLifetime = (time.Second) * 10

	// Length of the random value to be set on the leader lock. This
	// value will vary for each of the nodes/pods of the FS service. It
	// provides a way for the node to identify itself as the owner of the
	// lock before releasing it.
	randomStringLength = 10
)

var leaderLockValue = ""

// LUA script to release the leader lock. It checks to see if the value set for
// the leader lock key is the same random value determined for this node. If so,
// it proceeds to release the lock by deleting the key. If not, it fails to
// release the leader lock.
var releaseLeaderLockScript = redis.NewScript(`
if redis.call("get",KEYS[1]) == ARGV[1] then
  return redis.call("del",KEYS[1])
else
  return 0
end
`)

// AcquireLeaderLock attempts to acquire a lock for this FS node by writing
// a random string value to the leader lock key in the FS Redis cache. If
// the leader lock key is already set in the cache, the lock cannot be
// acquired. The caller can sleep for a while and retry acquiring the leader
// lock.
func AcquireLeaderLock() bool {
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	// If not already set, assign a new random value to be set for the
	// leader lock.
	if leaderLockValue == "" {
		leaderLockValue = common.NewRandomString(randomStringLength)
	}

	// Set key to hold string value if key does not exist. In that case,
	// it is equal to SET. When key already holds a value, no operation
	// is performed. SETNX is short for "SET if Not eXists".
	acquired, err := cacheClient.SetNX(ctx, leaderLockName, leaderLockValue,
		LeaderLockLifetime).Result()
	if (err != nil) || (!acquired) {
		fsLogger.Error("Failed to acquire the leader lock!",
			zap.Bool("Lock acquisition status:", acquired),
			zap.Error(err),
		)
		return false
	}

	return true
}

// ReleaseLeaderLock releases the leader lock if it is currently held by this
// FS node. Ownership of the leader lock is determined by checking to see if the
// leader lock key in the FS redis cache is set to the expected random string for
// this node. If the key is set to any other value, the lock cannot be released
// since this node doesn't own it.
func ReleaseLeaderLock() {
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	// Release the leader lock.
	_, err := releaseLeaderLockScript.Run(ctx, cacheClient,
		[]string{leaderLockName}, leaderLockValue).Int()
	if err != nil {
		fsLogger.Error("Failed to execute the LUA script to release the leader lock!",
			zap.Error(err),
		)
	}
}
