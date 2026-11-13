// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

// MessagingService handles email, SMS, and WhatsApp dispatch.
type MessagingService struct {
	client *Client
}

// GetProcessor returns the messaging processor configuration.
func (s *MessagingService) GetProcessor() (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("GET", "/api/v1/messaging/processor", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateProcessor updates the messaging processor configuration.
func (s *MessagingService) UpdateProcessor(input map[string]interface{}) (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("PUT", "/api/v1/messaging/processor", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SendEmail sends a transactional email (POST /api/v1/messaging/email/send).
//
// Keys of input, as the route reads them:
//   - "from" (required): an address or "Name <address>", at most 320 characters
//   - "to" (required, []string), "cc" / "bcc" ([]string, up to 10 each): one
//     address or "Name <address>" per entry
//   - "subject", "html" (required); "text"; "callerRef" (at most 200
//     characters, stored with the email's message log entries);
//     "attachments" ([]map with a non-empty "filename" and base64 "content")
//   - "replyTo": a string, or a []string of 1 to 5, each at most 320 characters
//   - "templateId", with "variables" (map[string]string) or none: a built-in,
//     platform or your organisation's template (another's or an unknown id:
//     404); its subject, HTML and text replace "subject" / "html" / "text"
//
// The server checks every limit and answers 400 or 404 (an *APIError).
func (s *MessagingService) SendEmail(input map[string]interface{}) (*SendMessageResult, error) {
	var result SendMessageResult
	err := s.client.doRequestTyped("POST", "/api/v1/messaging/email/send", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SendSMS sends an SMS message.
func (s *MessagingService) SendSMS(to, message string) (*SendMessageResult, error) {
	var result SendMessageResult
	err := s.client.doRequestTyped("POST", "/api/v1/messaging/sms", map[string]interface{}{
		"to": to, "message": message,
	}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SendWhatsApp dispatches a WhatsApp send.
//
// PlatformXe defaults to click-to-chat mode: when no Meta credentials are
// configured server-side, the response carries clickToChatUrl + deepLink
// instead of a messageId. The calling app surfaces those URLs to the operator
// (e.g. an "Open WhatsApp" button) and the operator's own WhatsApp Desktop /
// mobile app handles the send. When Meta credentials are configured, the call
// routes through Meta's Graph API and returns a messageId instead.
func (s *MessagingService) SendWhatsApp(input map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/messaging/whatsapp", input, nil)
}

// BuildWhatsAppLink builds a click-to-chat URL pair (wa.me + whatsapp://) for
// the given recipient + optional pre-populated message. Pure transform, no
// audit log. Use SendWhatsApp() if you want the call recorded.
func (s *MessagingService) BuildWhatsAppLink(to string, message string) (map[string]interface{}, error) {
	body := map[string]interface{}{"to": to}
	if message != "" {
		body["message"] = message
	}
	return s.client.doRequest("POST", "/api/v1/messaging/whatsapp/link", body, nil)
}

// WhatsAppHealth runs the WhatsApp diagnostic probe. Returns the active
// operating mode (click_to_chat vs meta_cloud_api), per-check results, and
// the calling organisation's own most recent FAILED rows from the
// transactional message log (recipients masked).
func (s *MessagingService) WhatsAppHealth() (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/messaging/whatsapp/health", nil, nil)
}

// EmailHealth returns the health status of the email service.
func (s *MessagingService) EmailHealth() (*EmailHealthResponse, error) {
	var result EmailHealthResponse
	err := s.client.doRequestTyped("GET", "/api/v1/messaging/email/health", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SMSHealth returns the health status of the SMS service.
func (s *MessagingService) SMSHealth() (*SmsHealthResponse, error) {
	var result SmsHealthResponse
	err := s.client.doRequestTyped("GET", "/api/v1/messaging/sms/health", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QueueStats returns email queue statistics.
func (s *MessagingService) QueueStats() (*QueueStats, error) {
	var result QueueStats
	err := s.client.doRequestTyped("GET", "/api/v1/messaging/queue/stats", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ProcessQueue triggers processing of the email retry queue (20 due emails).
// Platform keys only: the queue holds every organisation's emails, so a key
// with an organisation gets a 403 FORBIDDEN *APIError. The queue processes on
// its own schedule; tenants never need this.
func (s *MessagingService) ProcessQueue() (*QueueProcessResult, error) {
	var result QueueProcessResult
	err := s.client.doRequestTyped("POST", "/api/v1/messaging/queue/process", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
