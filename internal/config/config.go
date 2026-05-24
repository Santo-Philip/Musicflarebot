package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type BotConfig struct {
	ApiId             int32
	ApiHash           string
	Token             string
	SessionStrings    []string
	SessionType       string
	DatabaseUrl       string
	ApiUrl            string
	ApiKey            string
	OwnerId           int64
	LoggerId          int64
	Proxy             string
	DefaultService    string
	MaxFileSize       int64
	SongDurationLimit int64
	DownloadsDir      string
	SupportGroup      string
	SupportChannel    string
	DEVS              []int64
	CookiesPath       []string
	StartImg          string
	Port              string
	AutoLeave         bool
	LogLevel          string
}

var Conf *BotConfig

func LoadConfig() {
	loadDotEnv(".env.local", ".env")

	Conf = &BotConfig{
		ApiId:             int32(getEnvInt64("API_ID", 0)),
		ApiHash:           getEnvStr("API_HASH", ""),
		Token:             getEnvStr("TOKEN", ""),
		SessionType:       getEnvStr("SESSION_TYPE", "gogram"),
		DatabaseUrl:       getEnvStr("DATABASE_URL", ""),
		ApiUrl:            getEnvStr("API_URL", "https://beta.fallenapi.fun"),
		ApiKey:            getEnvStr("API_KEY", ""),
		OwnerId:           getEnvInt64("OWNER_ID", 0),
		Proxy:             getEnvStr("PROXY", ""),
		MaxFileSize:       getEnvInt64("MAX_FILE_SIZE", 500) * 1024 * 1024,
		SongDurationLimit: getEnvInt64("SONG_DURATION_LIMIT", 3600),
		DownloadsDir:      getEnvStr("DOWNLOADS_DIR", "downloads"),
		StartImg:          getEnvStr("START_IMG", "https://i.pinimg.com/originals/f5/42/4f/f5424f3a6e93438f441ce5d7a1c19e2c.mp4"),
		Port:              getEnvStr("PORT", "6060"),
		AutoLeave:         getEnvBool("AUTO_LEAVE", false),
		LogLevel:          getEnvStr("LOG_LEVEL", "info"),
		SessionStrings:    getSessionStrings("STRING", 10),
	}

	if Conf.OwnerId != 0 {
		Conf.DEVS = append([]int64{Conf.OwnerId}, Conf.DEVS...)
	}

	Conf.loadCookies()
	Conf.validate()
	Conf.ensureDirs()
}

func getEnvStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		return strings.EqualFold(v, "true") || v == "1"
	}
	return def
}

func getSessionStrings(prefix string, max int) []string {
	var s []string
	for i := 1; i <= max; i++ {
		if v := os.Getenv(fmt.Sprintf("%s%d", prefix, i)); v != "" {
			s = append(s, v)
		}
	}
	if v := os.Getenv(prefix); v != "" {
		s = append(s, v)
	}
	return s
}

func loadDotEnv(paths ...string) {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			if i := strings.IndexByte(line, '='); i != -1 {
				k := strings.TrimSpace(line[:i])
				v := strings.TrimSpace(line[i+1:])
				if len(v) > 1 && (v[0] == '"' || v[0] == '\'') && v[0] == v[len(v)-1] {
					v = v[1 : len(v)-1]
				}
				if os.Getenv(k) == "" {
					os.Setenv(k, v)
				}
			}
		}
	}
}

func (c *BotConfig) validate() {
	req := map[string]string{
		"API_ID": strconv.Itoa(int(c.ApiId)), "API_HASH": c.ApiHash,
		"TOKEN": c.Token, "DATABASE_URL": c.DatabaseUrl,
		"OWNER_ID": strconv.FormatInt(c.OwnerId, 10),
	}
	for name, val := range req {
		if val == "" || val == "0" {
			slog.Error("missing required config", "key", name)
			panic(fmt.Sprintf("missing required config: %s", name))
		}
	}
}

func (c *BotConfig) ensureDirs() {
	for _, d := range []string{c.DownloadsDir, "src/cookies"} {
		os.MkdirAll(d, 0755)
	}
}

func (c *BotConfig) loadCookies() {
	cd := filepath.Join("src", "cookies")
	entries, err := os.ReadDir(cd)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			c.CookiesPath = append(c.CookiesPath, filepath.Join(cd, e.Name()))
		}
	}
}
