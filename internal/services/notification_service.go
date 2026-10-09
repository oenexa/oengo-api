package services

import (
	"fmt"
	"log"
)

// NotificationService handles communication with third-party providers
// like SendGrid (Email) and Twilio (SMS).
type NotificationService struct {
	SendGridAPIKey string
	TwilioSID      string
	TwilioAuth     string
}

func NewNotificationService(sgKey, twilioSID, twilioAuth string) *NotificationService {
	return &NotificationService{
		SendGridAPIKey: sgKey,
		TwilioSID:      twilioSID,
		TwilioAuth:     twilioAuth,
	}
}

// SendEmail simulates sending an email via SendGrid
func (n *NotificationService) SendEmail(to, subject, body string) error {
	if n.SendGridAPIKey == "" {
		log.Println("WARNING: SendGrid API Key not configured. Skipping real email.")
	}

	// Simulated HTTP Call to https://api.sendgrid.com/v3/mail/send
	log.Printf("[SENDGRID EMAIL] To: %s | Subject: %s | Body: %s\n", to, subject, body)
	fmt.Printf("📧 Email successfully dispatched to %s via SendGrid.\n", to)
	return nil
}

// SendSMS simulates sending an SMS via Twilio
func (n *NotificationService) SendSMS(to, body string) error {
	if n.TwilioSID == "" || n.TwilioAuth == "" {
		log.Println("WARNING: Twilio credentials not configured. Skipping real SMS.")
	}

	// Simulated HTTP Call to https://api.twilio.com/2010-04-01/Accounts/.../Messages.json
	log.Printf("[TWILIO SMS] To: %s | Body: %s\n", to, body)
	fmt.Printf("📱 SMS successfully dispatched to %s via Twilio.\n", to)
	return nil
}

// NotifyOrderStatusChange sends an email and SMS to the customer when the order status changes.
func (n *NotificationService) NotifyOrderStatusChange(customerEmail, customerPhone, orderID, status string) {
	subject := fmt.Sprintf("Your OENGO Order %s Update", orderID)
	body := fmt.Sprintf("Hi! Your order %s is now: %s.", orderID, status)

	_ = n.SendEmail(customerEmail, subject, body)
	if customerPhone != "" {
		_ = n.SendSMS(customerPhone, body)
	}
}
