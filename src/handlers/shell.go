package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"musicflarebot/config"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func runShellCommand(cmd string, timeout time.Duration) (string, string, int) {
	var shell string
	var args []string

	if runtime.GOOS == "windows" {
		shell = "cmd"
		args = []string{"/C", cmd}
	} else {
		shell = "bash"
		args = []string{"-c", cmd}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := exec.CommandContext(ctx, shell, args...)

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Sprintf("Command timed out after %v seconds", timeout.Seconds()), -1
	}

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}

	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), exitCode
}

func shellRunner(m *tg.NewMessage) error {
	args := strings.TrimSpace(Args(m))
	if args == "" {
		_, _ = m.Reply("Usage: /sh cmd")
		return nil
	}

	msg, err := m.Reply("Running...")
	if err != nil {
		return nil
	}

	commands := strings.Split(args, "\n")
	var outputParts []string

	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		stdout, stderr, code := runShellCommand(cmd, 300*time.Second)
		part := fmt.Sprintf("<b>Command:</b> <code>%s</code>\n", cmd)
		if stdout != "" {
			part += fmt.Sprintf("<b>Output:</b>\n<pre>%s</pre>\n", stdout)
		}

		if stderr != "" {
			part += fmt.Sprintf("<b>Error:</b>\n<pre>%s</pre>\n", stderr)
		}
		part += fmt.Sprintf("<b>Exit Code:</b> <code>%d</code>\n", code)
		outputParts = append(outputParts, part)
	}

	finalOutput := strings.Join(outputParts, "\n")
	if strings.TrimSpace(finalOutput) == "" {
		finalOutput = "<b>📭 No output was returned</b>"
	}

	if len(finalOutput) <= 3500 {
		_, _ = msg.Edit(finalOutput, &tg.SendOptions{ParseMode: "HTML"})
		return nil
	}

	file := filepath.Join(config.Conf.DownloadsDir, fmt.Sprintf("%d.txt", time.Now().UnixNano()))
	if err := os.WriteFile(file, []byte(finalOutput), 0644); err != nil {
		_, _ = msg.Edit(fmt.Sprintf("Failed to write output: %v", err))
		return nil
	}
	defer os.Remove(file)

	_, err = msg.ReplyMedia(file, &tg.MediaOptions{
		ForceDocument: true,
	})

	if err != nil {
		_, _ = msg.Edit("Error: " + err.Error())
		return nil
	}

	return nil
}

func shellCommand(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	return shellRunner(m)
}
