package cache

import (
	"sync"

	"musicflarebot/internal/types"
)

type ChatData struct {
	Queue []*types.CachedTrack
}

type ChatCacher struct {
	mu         sync.RWMutex
	chatCache  map[int64]*ChatData
	eqPresets  map[int64]string
}

var ChatCache = newChatCacher()

func newChatCacher() *ChatCacher {
	return &ChatCacher{
		chatCache: make(map[int64]*ChatData),
		eqPresets: make(map[int64]string),
	}
}

func (c *ChatCacher) GetEQPreset(chatID int64) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.eqPresets[chatID]
}

func (c *ChatCacher) SetEQPreset(chatID int64, preset string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eqPresets[chatID] = preset
}

func (c *ChatCacher) getOrCreate(chatID int64) *ChatData {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, ok := c.chatCache[chatID]
	if !ok {
		data = &ChatData{}
		c.chatCache[chatID] = data
	}
	return data
}

func (c *ChatCacher) get(chatID int64) *ChatData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.chatCache[chatID]
	if !ok {
		return nil
	}
	return data
}

func (c *ChatCacher) AddSong(chatID int64, song *types.CachedTrack) int {
	data := c.getOrCreate(chatID)
	data.Queue = append(data.Queue, song)
	return len(data.Queue)
}

func (c *ChatCacher) AddSongs(chatID int64, songs []*types.CachedTrack) int {
	data := c.getOrCreate(chatID)
	data.Queue = append(data.Queue, songs...)
	return len(data.Queue)
}

func (c *ChatCacher) GetPlayingTrack(chatID int64) *types.CachedTrack {
	data := c.get(chatID)
	if data == nil || len(data.Queue) == 0 {
		return nil
	}
	return data.Queue[0]
}

func (c *ChatCacher) GetUpcomingTrack(chatID int64) *types.CachedTrack {
	data := c.get(chatID)
	if data == nil || len(data.Queue) < 2 {
		return nil
	}
	return data.Queue[1]
}

func (c *ChatCacher) RemoveCurrentSong(chatID int64) *types.CachedTrack {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, ok := c.chatCache[chatID]
	if !ok || len(data.Queue) == 0 {
		return nil
	}

	removed := data.Queue[0]
	data.Queue[0] = nil
	data.Queue = data.Queue[1:]
	return removed
}

func (c *ChatCacher) RemoveTrack(chatID int64, index int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, ok := c.chatCache[chatID]
	if !ok || index < 0 || index >= len(data.Queue) {
		return false
	}

	q := data.Queue
	copy(q[index:], q[index+1:])
	q[len(q)-1] = nil
	data.Queue = q[:len(q)-1]
	return true
}

func (c *ChatCacher) IsActive(chatID int64) bool {
	data := c.get(chatID)
	return data != nil && len(data.Queue) > 0
}

func (c *ChatCacher) ClearChat(chatID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if data, ok := c.chatCache[chatID]; ok {
		for i := range data.Queue {
			data.Queue[i] = nil
		}
		delete(c.chatCache, chatID)
	}
}

func (c *ChatCacher) GetQueueLength(chatID int64) int {
	data := c.get(chatID)
	if data == nil {
		return 0
	}
	return len(data.Queue)
}

func (c *ChatCacher) GetLoopCount(chatID int64) int {
	data := c.get(chatID)
	if data == nil || len(data.Queue) == 0 {
		return 0
	}
	return data.Queue[0].Loop
}

func (c *ChatCacher) SetLoopCount(chatID int64, loop int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, ok := c.chatCache[chatID]
	if !ok || len(data.Queue) == 0 {
		return false
	}
	data.Queue[0].Loop = loop
	return true
}

func (c *ChatCacher) GetQueue(chatID int64) []*types.CachedTrack {
	data := c.get(chatID)
	if data == nil || len(data.Queue) == 0 {
		return nil
	}
	return append([]*types.CachedTrack(nil), data.Queue...)
}

func (c *ChatCacher) GetActiveChats() []int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	active := make([]int64, 0, len(c.chatCache))
	for chatID, data := range c.chatCache {
		if len(data.Queue) > 0 {
			active = append(active, chatID)
		}
	}
	return active
}

func (c *ChatCacher) GetTrackIfExists(chatID int64, trackID string) *types.CachedTrack {
	data := c.get(chatID)
	if data == nil {
		return nil
	}
	for _, t := range data.Queue {
		if t.TrackID == trackID {
			return t
		}
	}
	return nil
}
