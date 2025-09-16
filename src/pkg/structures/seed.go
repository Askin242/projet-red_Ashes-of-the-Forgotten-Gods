package structures

import (
	"hash/fnv"
	"math/rand"
	"sync"
)

var (
	globalRNG   *rand.Rand
	rngMutex    sync.RWMutex
	isSeeded    bool
	baseSeed    string
	currentSeed int64
)

func InitializeSeed(seedStr string) {
	rngMutex.Lock()
	defer rngMutex.Unlock()

	baseSeed = seedStr
	currentSeed = stringToSeed(seedStr)
	source := rand.NewSource(currentSeed)
	globalRNG = rand.New(source)
	isSeeded = true
}

func InitializeFromCurrentSeed(seedStr string, current int64) {
	rngMutex.Lock()
	defer rngMutex.Unlock()

	baseSeed = seedStr
	currentSeed = current
	source := rand.NewSource(currentSeed)
	globalRNG = rand.New(source)
	isSeeded = true
}

func GetRNG() *rand.Rand {
	rngMutex.Lock()
	defer rngMutex.Unlock()

	if !isSeeded {
		source := rand.NewSource(1)
		return rand.New(source)
	}

	currentSeed = globalRNG.Int63()

	return globalRNG
}

func GetCurrentSeedState() int64 {
	rngMutex.RLock()
	defer rngMutex.RUnlock()
	return currentSeed
}

func GetBaseSeed() string {
	rngMutex.RLock()
	defer rngMutex.RUnlock()
	return baseSeed
}

func IsSeeded() bool {
	rngMutex.RLock()
	defer rngMutex.RUnlock()
	return isSeeded
}

func stringToSeed(s string) int64 {
	if s == "" {
		return 1
	}

	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func GetSeedValue(seedStr string) int64 {
	return stringToSeed(seedStr)
}

func RefreshSeedState() {
	rngMutex.Lock()
	defer rngMutex.Unlock()
	if isSeeded && globalRNG != nil {
		currentSeed = globalRNG.Int63()
	}
}
