package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"musicflarebot/internal/config"
	"musicflarebot/internal/database"
	"musicflarebot/internal/logger"
	"musicflarebot/src"
	"musicflarebot/src/handlers"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func main() {
	config.LoadConfig()

	logger.Init(config.Conf.LogLevel)

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05"))
			}
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line))
			}
			return a
		},
	})))

	go func() {
		if err := http.ListenAndServe("0.0.0.0:"+config.Conf.Port, nil); err != nil {
			slog.Info("pprof server error", "error", err)
		}
	}()

	if err := database.Connect(config.Conf.DatabaseUrl); err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	client, err := tg.NewClient(tg.ClientConfig{
		AppID:     config.Conf.ApiId,
		AppHash:   config.Conf.ApiHash,
		ParseMode: "HTML",
	})
	if err != nil {
		slog.Error("tg.NewClient error", "error", err)
		os.Exit(1)
	}

	if err := client.Connect(); err != nil {
		slog.Error("client.Connect() error", "error", err)
		os.Exit(1)
	}

	if err := client.LoginBot(config.Conf.Token); err != nil {
		slog.Error("client.LoginBot() error", "error", err)
		os.Exit(1)
	}

	if err := src.Init(client); err != nil {
		slog.Error("src.Init error", "error", err)
		os.Exit(1)
	}

	handlers.LoadModules(client)

	me := client.Me()
	username := ""
	if me != nil {
		username = me.Username
	}

	slog.Info("Bot started", "username", username, "id", me.ID)
	client.SendMessage(config.Conf.LoggerId, "The bot has started!", nil)
	client.Idle()

	slog.Info("The bot is shutting down...")
	vc.Calls.StopAllClients()
}
