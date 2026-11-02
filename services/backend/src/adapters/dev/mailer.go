package dev

import (
	"context"
	"log/slog"
)

// Mailer logs emails instead of sending them.
type Mailer struct {
	logger *slog.Logger
}

func NewMailer(logger *slog.Logger) *Mailer {
	return &Mailer{logger: logger}
}

func (m *Mailer) SendLoginLink(ctx context.Context, to, link string) error {
	m.logger.InfoContext(ctx, "login link", "to", to, "link", link)
	return nil
}
