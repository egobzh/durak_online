package notifier

import (
	"context"
	"fmt"
	"log"
)

type ConsoleResetLinkSender struct {
	baseURL string
}

func NewConsoleResetLinkSender(baseURL string) *ConsoleResetLinkSender {
	if baseURL == "" {
		baseURL = "http://localhost:8080/api/v1/auth/reset-password"
	}
	return &ConsoleResetLinkSender{baseURL: baseURL}
}

func (s *ConsoleResetLinkSender) SendPasswordResetLink(ctx context.Context, email, token string) error {
	resetLink := fmt.Sprintf("%s?token=%s", s.baseURL, token)

	log.Printf("\n=======================================================\n")
	log.Printf("[PASSWORD RESET NOTIFICATION]")
	log.Printf("To: %s", email)
	log.Printf("Reset Token: %s", token)
	log.Printf("Reset Link:  %s", resetLink)
	log.Printf("Expires in:  30 minutes")
	log.Printf("=======================================================\n")

	return nil
}
