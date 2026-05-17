package vc

import (
	"musicflarebot/config"
	"musicflarebot/src/vc/sessions"
	"musicflarebot/src/vc/ubot"
	"fmt"
	"log/slog"

	tg "github.com/amarnathcjd/gogram/telegram"
)

// StartClient initializes a new userbot client and adds it to the pool of available assistants.
// It authenticates with Telegram using the provided API ID, API hash, and session string.
// The session type is determined by the configuration (pyrogram, telethon, or gogram).
func (c *TelegramCalls) StartClient(apiID int32, apiHash, stringSession string) (*ubot.Context, error) {
	c.mu.Lock()
	clientIndex := len(c.uBContext)
	c.mu.Unlock()

	clientName := fmt.Sprintf("client%d", clientIndex)

	var sess *tg.Session
	var err error

	clientConfig := tg.ClientConfig{
		AppID:         apiID,
		AppHash:       apiHash,
		MemorySession: true,
		SessionName:   clientName,
		FloodHandler:  handleFlood,
		LogLevel:      tg.InfoLevel,
	}

	switch config.Conf.SessionType {
	case "telethon":
		sess, err = sessions.DecodeTelethonSessionString(stringSession)
		if err != nil {
			return nil, fmt.Errorf("failed to decode telethon session string for %s: %w", clientName, err)
		}
		clientConfig.StringSession = sess.Encode()
	case "pyrogram":
		sess, err = sessions.DecodePyrogramSessionString(stringSession)
		if err != nil {
			return nil, fmt.Errorf("failed to decode pyrogram session string for %s: %w", clientName, err)
		}
		clientConfig.StringSession = sess.Encode()
	case "gogram":
		clientConfig.StringSession = stringSession
	default:
		return nil, fmt.Errorf("unsupported session type: %s", config.Conf.SessionType)
	}

	mtProto, err := tg.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create the MTProto client: %w", err)
	}

	if err = mtProto.Start(); err != nil {
		return nil, fmt.Errorf("failed to start the client: %w", err)
	}

	me := mtProto.Me()
	if me.Bot {
		_ = mtProto.Stop()
		return nil, fmt.Errorf("the client %s is a bot", clientName)
	}

	appConfig, err := mtProto.HelpGetAppConfig(0)
	if err != nil {
		logger.Warn("[TelegramCalls] failed to fetch app config", "client", clientName, "error", err)
	} else {
		isFreeze := false
		if cfg, ok := appConfig.(*tg.HelpAppConfigObj); ok {
			if cfgObj, ok := cfg.Config.(*tg.JsonObject); ok {
				for _, entry := range cfgObj.Value {
					if entry != nil && entry.Key == "freeze_since_date" {
						isFreeze = true
						break
					}
				}
			}
		}

		if isFreeze {
			logger.Warn("[TelegramCalls] The client is frozen and cannot be used for voice calls", "client", clientName, "id", me.ID, "username", me.Username)
			_ = mtProto.Stop()
			return nil, nil
		}
	}

	call, err := ubot.NewInstance(mtProto)
	if err != nil {
		_ = mtProto.Stop()
		return nil, fmt.Errorf("failed to create the ubot instance: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.uBContext[clientIndex] = call
	c.clients[clientIndex] = mtProto
	c.clientsBySession[stringSession] = clientIndex

	logger.Info("[TelegramCalls] Client started", "client", clientName, "id", me.ID, "username", me.Username)
	return call, nil
}

// StopClient stops a specific userbot client identified by its session string.
func (c *TelegramCalls) StopClient(sessionStr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx, ok := c.clientsBySession[sessionStr]
	if !ok {
		return fmt.Errorf("client for session string not found")
	}

	if cl, ok := c.clients[idx]; ok {
		slog.Info("[TelegramCalls] Stopping client", "index", idx)
		if err := cl.Stop(); err != nil {
			slog.Warn("[TelegramCalls] Error stopping client", "index", idx, "error", err)
		}
		delete(c.clients, idx)
	}

	if ctx, ok := c.uBContext[idx]; ok {
		ctx.Close()
		delete(c.uBContext, idx)
	}

	delete(c.clientsBySession, sessionStr)
	return nil
}

// StopAllClients gracefully stops all active userbot clients and their associated voice calls.
func (c *TelegramCalls) StopAllClients() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, call := range c.uBContext {
		call.Close()
	}

	for idx, client := range c.clients {
		slog.Info("[TelegramCalls] Stopping the client", "index", idx)
		_ = client.Stop()
	}
}
