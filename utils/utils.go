package utils

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"

	swagger "github.com/infiotinc/netskopebwan-go-client"
)

const NAT = "1:1_NAT"
const PORT_FORWARD = "PORT_FORWARD"

var Mutex = KeyMutex{}

type KeyMutex struct {
	m  map[string]*sync.Mutex
	mu sync.Mutex
}

func GetExistingNat(natRules []swagger.InboundNatRule,
	natEntry swagger.InboundNatRule) (index int) {
	for index, nat := range natRules {
		if nat.Name == natEntry.Name {
			return index
		}
	}
	return -1
}

func (ctx *KeyMutex) Get(key string) *sync.Mutex {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.m == nil {
		ctx.m = map[string]*sync.Mutex{}
	}
	if l, ok := ctx.m[key]; ok {
		return l
	} else {
		ctx.m[key] = &sync.Mutex{}
		return ctx.m[key]
	}
}

func Hash(input interface{}) string {
	jsonBytes, _ := json.Marshal(input)
	return fmt.Sprintf("%x", md5.Sum(jsonBytes))
}
