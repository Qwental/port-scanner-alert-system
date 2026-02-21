package notifier

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type TelegramNotifier struct {
	token  string
	chatID int64
	log    *zap.SugaredLogger
}

func NewTelegramNotifier(token string, chatID int64, log *zap.SugaredLogger) *TelegramNotifier {
	return &TelegramNotifier{
		token:  token,
		chatID: chatID,
		log:    log,
	}
}

const telegramMaxLen = 4096

func (t *TelegramNotifier) Send(message string) error {
	runes := []rune(message)
	for i := 0; i < len(runes); i += telegramMaxLen {
		end := i + telegramMaxLen
		if end > len(runes) {
			end = len(runes)
		}
		if err := t.sendChunk(string(runes[i:end])); err != nil {
			return err
		}
	}
	return nil
}

func (t *TelegramNotifier) sendChunk(text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)

	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id":    {fmt.Sprintf("%d", t.chatID)},
		"text":       {text},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("telegram response parse failed: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("telegram API error: %s", result.Description)
	}

	t.log.Info("Telegram notification sent")
	return nil
}
