package fctrl

import "github.com/go-redis/redis"

type manager struct {
	runMode            int // bit encoding, default is 0, 0/1 determine single mode/redis mode
	redisMode          *int
	redisClient        *redis.Client
	redisClusterClient *redis.ClusterClient
}

func NewManager(productRules *ProductRules, options ...ManagerOption) error {
	mgr := &manager{}
	for _, optionFunc := range options {
		optionFunc(mgr)
	}
	globalManager = mgr
	return nil
}

type ManagerOption func(mgr *manager)

func ManagerOptionRedisClient(client *redis.Client) ManagerOption {
	return func(mgr *manager) {
		mgr.redisClient = client
		mgr.formatMode(runRedisMode, runRedisClient)
	}
}

func ManagerOptionRedisClusterClient(clusterClient *redis.ClusterClient) ManagerOption {
	return func(mgr *manager) {
		mgr.redisClusterClient = clusterClient
		mgr.formatMode(runRedisMode, runRedisClusterClient)
	}
}

// formatMode format run mdoe
func (mgr *manager) formatMode(runMode int, redisModes ...int) {
	mgr.runMode = runMode
	if len(redisModes) > 0 {
		redisMode := redisModes[0]
		mgr.redisMode = &redisMode
	}
}
