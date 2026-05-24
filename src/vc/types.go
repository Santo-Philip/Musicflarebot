package vc

import (
	"log/slog"
	"regexp"
	"sync"
	"time"

	"musicflarebot/internal/cache"
	"musicflarebot/src/vc/ubot"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var logger = slog.Default()
var urlRegex = regexp.MustCompile(`^https?://`)

// TelegramCalls manages the state and operations for voice calls, including userbots and the main bot client.
type TelegramCalls struct {
	mu               sync.RWMutex
	uBContext        map[int]*ubot.Context
	clients          map[int]*tg.Client
	bot              *tg.Client
	statusCache      *cache.Cache[string]
	inviteCache      *cache.Cache[string]
	clientsBySession map[string]int // session string -> client index
}

var (
	instance *TelegramCalls
	once     sync.Once
)

// getCalls returns the singleton instance of the TelegramCalls manager, ensuring that only one instance is created.
func getCalls() *TelegramCalls {
	once.Do(func() {
		instance = &TelegramCalls{
			uBContext:        make(map[int]*ubot.Context),
			clients:          make(map[int]*tg.Client),
			statusCache:      cache.NewCache[string](2 * time.Hour),
			inviteCache:      cache.NewCache[string](2 * time.Hour),
			clientsBySession: make(map[string]int),
		}
	})
	return instance
}

// Calls is the singleton instance of TelegramCalls, initialized lazily.
var Calls = getCalls()
