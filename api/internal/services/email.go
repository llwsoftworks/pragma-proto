package services

import (
	"fmt"

	"github.com/resendlabs/resend-go"
)

// EmailService sends transactional emails via Resend.
type EmailService struct {
	client    *resend.Client
	fromAddr  string
}

// NewEmailService creates an EmailService.
func NewEmailService(apiKey, fromAddr string) *EmailService {
	return &EmailService{
		client:   resend.NewClient(apiKey),
		fromAddr: fromAddr,
	}
}

// SendPasswordReset emails a password reset link to the user.
func (s *EmailService) SendPasswordReset(to, firstName, resetURL string) error {
	body := fmt.Sprintf(`<p>Hello %s,</p>
<p>You requested a password reset. Click the link below to set a new password. This link expires in 1 hour.</p>
<p><a href="%s">Reset Password</a></p>
<p>If you did not request this, please ignore this email.</p>`, firstName, resetURL)

	params := &resend.SendEmailRequest{
		From:    s.fromAddr,
		To:      []string{to},
		Subject: "Reset your password",
		Html:    body,
	}
	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("email: send password reset: %w", err)
	}
	return nil
}

// SendGradeUnlock notifies a student and their parents that grade access has been restored.
func (s *EmailService) SendGradeUnlock(to []string, studentName string) error {
	body := fmt.Sprintf(`<p>This is a notification that grade access has been restored for %s.</p>
<p>You can now log in to view your grades.</p>`, studentName)

	params := &resend.SendEmailRequest{
		From:    s.fromAddr,
		To:      to,
		Subject: "Grade access restored",
		Html:    body,
	}
	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("email: send grade unlock: %w", err)
	}
	return nil
}

// SendParentCommunication sends a teacher-authored message to a parent.
func (s *EmailService) SendParentCommunication(to, subject, htmlBody string) error {
	params := &resend.SendEmailRequest{
		From:    s.fromAddr,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}
	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("email: send parent communication: %w", err)
	}
	return nil
}

// SendReportCardReady notifies a parent that a report card is available.
func (s *EmailService) SendReportCardReady(to, studentName, period, downloadURL string) error {
	body := fmt.Sprintf(`<p>The report card for %s for the %s period is now available.</p>
<p><a href="%s">Download Report Card</a></p>`, studentName, period, downloadURL)

	params := &resend.SendEmailRequest{
		From:    s.fromAddr,
		To:      []string{to},
		Subject: fmt.Sprintf("Report card available — %s", period),
		Html:    body,
	}
	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("email: send report card ready: %w", err)
	}
	return nil
}
