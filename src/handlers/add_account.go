package handlers

import (
	"fmt"
	"log/slog"
	"musicflarebot/internal/config"
	"musicflarebot/internal/database"
	"musicflarebot/src/vc"
	"strconv"
	"strings"
	"sync"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

type accState int

const (
	stateNone accState = iota
	statePhone
	stateOTP
	state2FA
)

type accSession struct {
	State  accState
	Phone  string
	Client *tg.Client
	Hash   string
}

var (
	accMu       sync.Mutex
	accSessions = make(map[int64]*accSession)
)

func addaccHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	userID := m.SenderID()

	accMu.Lock()
	if _, exists := accSessions[userID]; exists {
		accMu.Unlock()
		_, _ = m.Reply("You already have an active login session. Use /cancel to abort.")
		return nil
	}
	accMu.Unlock()

	phoneMsg, err := m.Reply(
		"<b>Assistant Account Setup</b>\n\n" +
			"Please send your <b>phone number</b> in international format (e.g., +1234567890).\n" +
			"Use /cancel to abort.",
	)
	if err != nil {
		return err
	}

	accMu.Lock()
	accSessions[userID] = &accSession{
		State: statePhone,
	}
	accMu.Unlock()

	_ = phoneMsg
	return nil
}

func handleAddAccMessage(m *tg.NewMessage) error {
	userID := m.SenderID()

	accMu.Lock()
	session, exists := accSessions[userID]
	accMu.Unlock()

	if !exists {
		return nil
	}

	// m.Text is a method returning the message text; call it to obtain the string value.
	text := strings.TrimSpace(m.Text())

	switch session.State {
	case statePhone:
		return processPhone(m, session, text)
	case stateOTP:
		return processOTP(m, session, text)
	case state2FA:
		return process2FA(m, session, text)
	default:
		return nil
	}
}

func processPhone(m *tg.NewMessage, session *accSession, phone string) error {
	if !strings.HasPrefix(phone, "+") {
		_, _ = m.Reply("Invalid phone number. Please use international format (e.g., +1234567890).")
		return nil
	}

	client, err := tg.NewClient(tg.ClientConfig{
		AppID:         config.Conf.ApiId,
		AppHash:       config.Conf.ApiHash,
		ParseMode:     "HTML",
		MemorySession: true,
	})
	if err != nil {
		slog.Error("tg.NewClient error", "error", err)
		_, _ = m.Reply(fmt.Sprintf("Failed to create client: %s", err.Error()))

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	if err = client.Connect(); err != nil {
		slog.Error("client.Connect() error", "error", err)
		_, _ = m.Reply(fmt.Sprintf("Failed to connect: %s", err.Error()))

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	hash, err := client.SendCode(phone)
	if err != nil {
		slog.Error("client.SendCode error", "error", err)
		_, _ = m.Reply(fmt.Sprintf("Failed to send code: %s", err.Error()))

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	session.Phone = phone
	session.Client = client
	session.Hash = hash
	session.State = stateOTP

	_, _ = m.Reply("An OTP has been sent to your Telegram account. Please send the OTP code.")
	return nil
}

func processOTP(m *tg.NewMessage, session *accSession, code string) error {
	auth, err := session.Client.AuthSignIn(session.Phone, session.Hash, code, nil)
	if err != nil {
		if is2FAError(err) {
			session.State = state2FA
			_, _ = m.Reply("Two-factor authentication is enabled. Please send your 2FA password.")
			return nil
		}

		slog.Error("AuthSignIn error", "error", err)
		_, _ = m.Reply(fmt.Sprintf("OTP verification failed: %s", err.Error()))

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	_ = auth
	return finalizeAccount(m, session)
}

func process2FA(m *tg.NewMessage, session *accSession, password string) error {
	accPassword, err := session.Client.AccountGetPassword()
	if err != nil {
		slog.Error("AccountGetPassword error", "error", err)
		_, _ = m.Reply(fmt.Sprintf("Failed to get password info: %s", err.Error()))

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	inputPassword, err := tg.GetInputCheckPassword(password, accPassword)
	if err != nil {
		slog.Error("GetInputCheckPassword error", "error", err)
		_, _ = m.Reply(fmt.Sprintf("Failed to compute password hash: %s", err.Error()))

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	if _, err := session.Client.AuthCheckPassword(inputPassword); err != nil {
		slog.Error("AuthCheckPassword error", "error", err)
		_, _ = m.Reply(fmt.Sprintf("2FA password verification failed: %s", err.Error()))

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	return finalizeAccount(m, session)
}

func finalizeAccount(m *tg.NewMessage, session *accSession) error {
	sessionStr := session.Client.ExportSession()
	if sessionStr == "" {
		slog.Error("ExportSession returned empty string")
		_, _ = m.Reply("Failed to export session: empty session string")

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	me := session.Client.Me()
	userID := me.ID
	phone := me.Phone

	_ = session.Client.Stop()

	key := fmt.Sprintf("session_db_%d", time.Now().UnixNano())

	if err := database.SetSetting(key, sessionStr); err != nil {
		slog.Error("Failed to save session string", "error", err)
		_, _ = m.Reply(fmt.Sprintf("Failed to save session string: %s", err.Error()))

		accMu.Lock()
		delete(accSessions, m.SenderID())
		accMu.Unlock()
		return nil
	}

	config.Conf.SessionStrings = append(config.Conf.SessionStrings, sessionStr)

	accountNumber := len(config.Conf.SessionStrings)

	_, err := vc.Calls.StartClient(
		config.Conf.ApiId,
		config.Conf.ApiHash,
		sessionStr,
	)
	if err != nil {
		_, _ = m.Reply(fmt.Sprintf("Session saved but failed to start assistant: %s", err.Error()))
	} else {
		_, _ = m.Reply(
			fmt.Sprintf(
				"✅ <b>Assistant account added successfully!</b>\n\n<b>User ID:</b> <code>%d</code>\n<b>Phone:</b> <code>%s</code>\n<b>Account #:</b> <code>%d</code>",
				userID,
				phone,
				accountNumber,
			),
		)
	}

	accMu.Lock()
	delete(accSessions, m.SenderID())
	accMu.Unlock()

	return nil
}

func cancelAddAccHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	userID := m.SenderID()

	accMu.Lock()
	session, exists := accSessions[userID]
	if exists {
		if session.Client != nil {
			_ = session.Client.Stop()
		}
		delete(accSessions, userID)
	}
	accMu.Unlock()

	if exists {
		_, _ = m.Reply("Login process cancelled.")
	} else {
		_, _ = m.Reply("No active login process to cancel.")
	}

	return nil
}

func listAccHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	sessions := config.Conf.SessionStrings
	if len(sessions) == 0 {
		_, _ = m.Reply("No assistant accounts configured.")
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Assistant Accounts (%d)</b>\n\n", len(sessions)))
	for i := range sessions {
		sb.WriteString(fmt.Sprintf("<b>%d.</b> Account <code>%d</code>\n", i+1, i+1))
	}

	_, err := m.Reply(sb.String(), &tg.SendOptions{ParseMode: "HTML"})
	return err
}

func removeAccHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /removeacc [index]\nUse /listacc to see available indices.")
		return nil
	}

	index, err := strconv.Atoi(args)
	if err != nil || index < 1 || index > len(config.Conf.SessionStrings) {
		_, _ = m.Reply(fmt.Sprintf("Invalid index. Use a number between 1 and %d.", len(config.Conf.SessionStrings)))
		return nil
	}

	sessions := config.Conf.SessionStrings
	removed := sessions[index-1]
	config.Conf.SessionStrings = append(sessions[:index-1], sessions[index:]...)

	_ = vc.Calls.StopClient(removed)

	_ = database.DeleteAllSessionKeys()
	for i, s := range config.Conf.SessionStrings {
		_ = database.SetSetting(fmt.Sprintf("session_db_%d", time.Now().UnixNano()+int64(i)), s)
	}

	_, _ = m.Reply(fmt.Sprintf("Account <b>%d</b> has been removed.", index))
	return nil
}

func is2FAError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "SESSION_PASSWORD_NEEDED") ||
		strings.Contains(errStr, "2FA") ||
		strings.Contains(errStr, "two-factor") ||
		strings.Contains(errStr, "password needed")
}
