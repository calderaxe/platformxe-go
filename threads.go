// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

// ThreadsService handles Caldera Threads (Contextual Messaging) API calls.
type ThreadsService struct {
	client *Client
}

// ── Input Types ─────────────────────────────────────────────────────────────

// CreateThreadChannelInput is the input for creating a channel.
type CreateThreadChannelInput struct {
	Slug              string                 `json:"slug"`
	DisplayName       string                 `json:"displayName"`
	EntityType        string                 `json:"entityType"`
	ParticipantRoles  []string               `json:"participantRoles"`
	DefaultVisibility []string               `json:"defaultVisibility"`
	LifecycleRules    map[string]interface{} `json:"lifecycleRules,omitempty"`
	EscalationConfig  map[string]interface{} `json:"escalationConfig,omitempty"`
	WebhookURL        string                 `json:"webhookUrl,omitempty"`
}

// UpdateThreadChannelInput is the input for updating a channel.
type UpdateThreadChannelInput struct {
	DisplayName       *string                `json:"displayName,omitempty"`
	ParticipantRoles  []string               `json:"participantRoles,omitempty"`
	DefaultVisibility []string               `json:"defaultVisibility,omitempty"`
	LifecycleRules    map[string]interface{} `json:"lifecycleRules,omitempty"`
	EscalationConfig  map[string]interface{} `json:"escalationConfig,omitempty"`
	WebhookURL        *string                `json:"webhookUrl,omitempty"`
	IsActive          *bool                  `json:"isActive,omitempty"`
}

// ThreadParticipant is a participant in a thread creation request.
type ThreadParticipant struct {
	Role        string `json:"role"`
	ExternalID  string `json:"externalId"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
}

// CreateThreadInput is the input for creating a thread.
type CreateThreadInput struct {
	ChannelSlug  string                 `json:"channelSlug"`
	EntityID     string                 `json:"entityId"`
	Subject      string                 `json:"subject,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Participants []ThreadParticipant    `json:"participants"`
}

// SendMessageInput is the input for sending a message.
type SendMessageInput struct {
	SenderExternalID string   `json:"senderExternalId"`
	SenderRole       string   `json:"senderRole"`
	Content          string   `json:"content"`
	Visibility       []string `json:"visibility,omitempty"`
}

// EntityEventInput is the input for forwarding entity status changes.
type EntityEventInput struct {
	ChannelSlug string `json:"channelSlug"`
	EntityID    string `json:"entityId"`
	Event       string `json:"event"`
	NewStatus   string `json:"newStatus,omitempty"`
}

// FlagMessageInput is the input for flagging a message.
type FlagMessageInput struct {
	Reason              string `json:"reason"`
	Note                string `json:"note,omitempty"`
	FlaggedByExternalID string `json:"flaggedByExternalId"`
	FlaggedByRole       string `json:"flaggedByRole"`
}

// ── Channels ─────────────────────────────────────────────────────────────────

// CreateChannel creates a new thread channel.
func (s *ThreadsService) CreateChannel(input CreateThreadChannelInput) (*ThreadChannel, error) {
	var result ThreadChannel
	err := s.client.doRequestTyped("POST", "/api/v1/threads/channels", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListChannels returns all channels for the organization.
func (s *ThreadsService) ListChannels() (*ChannelListResponse, error) {
	var result ChannelListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/threads/channels", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateChannel updates a channel configuration.
func (s *ThreadsService) UpdateChannel(channelID string, input UpdateThreadChannelInput) (*ThreadChannel, error) {
	var result ThreadChannel
	err := s.client.doRequestTyped("PATCH", "/api/v1/threads/channels/"+channelID, input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEscalationConfig returns the escalation configuration for a channel.
func (s *ThreadsService) GetEscalationConfig(channelID string) (*EscalationConfig, error) {
	var result EscalationConfig
	err := s.client.doRequestTyped("GET", "/api/v1/threads/channels/"+channelID+"/escalation", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SetEscalationConfig sets the escalation configuration for a channel.
func (s *ThreadsService) SetEscalationConfig(channelID string, config map[string]interface{}) (*EscalationConfig, error) {
	var result EscalationConfig
	err := s.client.doRequestTyped("PUT", "/api/v1/threads/channels/"+channelID+"/escalation", config, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Threads ──────────────────────────────────────────────────────────────────

// CreateThread creates a new thread with initial participants.
func (s *ThreadsService) CreateThread(input CreateThreadInput) (*Thread, error) {
	var result Thread
	err := s.client.doRequestTyped("POST", "/api/v1/threads", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetThread returns a thread by ID with participants.
func (s *ThreadsService) GetThread(threadID string) (*Thread, error) {
	var result Thread
	err := s.client.doRequestTyped("GET", "/api/v1/threads/"+threadID, nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListThreads returns threads matching the given filters.
func (s *ThreadsService) ListThreads(params map[string]string) (*ThreadListResponse, error) {
	var result ThreadListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/threads", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CloseThread closes a thread.
func (s *ThreadsService) CloseThread(threadID string, reason string) (*ThreadCloseResult, error) {
	var result ThreadCloseResult
	err := s.client.doRequestTyped("POST", "/api/v1/threads/"+threadID+"/close", map[string]string{"reason": reason}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ReopenThread reopens a closed thread.
func (s *ThreadsService) ReopenThread(threadID string) (*ThreadReopenResult, error) {
	var result ThreadReopenResult
	err := s.client.doRequestTyped("POST", "/api/v1/threads/"+threadID+"/reopen", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EntityEvent forwards an entity status change for lifecycle evaluation.
func (s *ThreadsService) EntityEvent(input EntityEventInput) (*EntityEventResult, error) {
	var result EntityEventResult
	err := s.client.doRequestTyped("POST", "/api/v1/threads/entity-event", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Messages ─────────────────────────────────────────────────────────────────

// SendMessage sends a message in a thread.
func (s *ThreadsService) SendMessage(threadID string, input SendMessageInput) (*ThreadMessage, error) {
	var result ThreadMessage
	err := s.client.doRequestTyped("POST", "/api/v1/threads/"+threadID+"/messages", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SendSystemMessage sends a system message in a thread.
func (s *ThreadsService) SendSystemMessage(threadID string, content string, visibility []string) (*ThreadMessage, error) {
	body := map[string]interface{}{"content": content, "visibility": visibility}
	var result ThreadMessage
	err := s.client.doRequestTyped("POST", "/api/v1/threads/"+threadID+"/messages/system", body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Read State & Inbox ───────────────────────────────────────────────────────

// MarkRead marks a thread as read for a participant.
func (s *ThreadsService) MarkRead(threadID string, externalID string, role string) (*MarkReadResult, error) {
	body := map[string]string{"participantExternalId": externalID, "participantRole": role}
	var result MarkReadResult
	err := s.client.doRequestTyped("POST", "/api/v1/threads/"+threadID+"/read", body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Inbox returns all threads for a participant with unread counts.
func (s *ThreadsService) Inbox(externalID string, role string, params map[string]string) (*InboxResponse, error) {
	if params == nil {
		params = map[string]string{}
	}
	params["externalId"] = externalID
	params["role"] = role
	var result InboxResponse
	err := s.client.doRequestTyped("GET", "/api/v1/threads/inbox", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UnreadCount returns total unread count for a participant.
func (s *ThreadsService) UnreadCount(externalID string, role string) (*UnreadCountResult, error) {
	params := map[string]string{"externalId": externalID, "role": role}
	var result UnreadCountResult
	err := s.client.doRequestTyped("GET", "/api/v1/threads/inbox/unread-count", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Flags & Escalation ───────────────────────────────────────────────────────

// FlagMessage flags a message for escalation.
func (s *ThreadsService) FlagMessage(threadID string, messageID string, input FlagMessageInput) (*FlagMessageResult, error) {
	var result FlagMessageResult
	err := s.client.doRequestTyped("POST", "/api/v1/threads/"+threadID+"/messages/"+messageID+"/flag", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EscalateThread escalates a thread (admin action).
func (s *ThreadsService) EscalateThread(threadID string, reason string, escalatedBy string, role string) (*EscalateThreadResult, error) {
	body := map[string]string{"reason": reason, "escalatedByExternalId": escalatedBy, "escalatedByRole": role}
	var result EscalateThreadResult
	err := s.client.doRequestTyped("POST", "/api/v1/threads/"+threadID+"/escalate", body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListFlags returns flags for a thread.
func (s *ThreadsService) ListFlags(threadID string, params map[string]string) (*FlagListResponse, error) {
	var result FlagListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/threads/"+threadID+"/flags", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Update Thread ───────────────────────────────────────────────────────────

// UpdateThread updates a thread's subject or metadata.
func (s *ThreadsService) UpdateThread(threadID string, input UpdateThreadInput) (*Thread, error) {
	var result Thread
	err := s.client.doRequestTyped("PATCH", "/api/v1/threads/"+threadID, input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Message Management ──────────────────────────────────────────────────────

// ListMessages returns messages in a thread with optional filters.
func (s *ThreadsService) ListMessages(threadID string, params map[string]string) (*MessageListResponse, error) {
	var result MessageListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/threads/"+threadID+"/messages", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EditMessage edits the content of a message.
func (s *ThreadsService) EditMessage(messageID string, content string) (*ThreadMessage, error) {
	var result ThreadMessage
	err := s.client.doRequestTyped("PATCH", "/api/v1/threads/messages/"+messageID, map[string]string{"content": content}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteMessage soft-deletes a message.
func (s *ThreadsService) DeleteMessage(messageID string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", "/api/v1/threads/messages/"+messageID, nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Participant Management ──────────────────────────────────────────────────

// AddParticipant adds a participant to a thread.
func (s *ThreadsService) AddParticipant(threadID string, input AddParticipantInput) (*ThreadParticipantInfo, error) {
	var result ThreadParticipantInfo
	err := s.client.doRequestTyped("POST", "/api/v1/threads/"+threadID+"/participants", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RemoveParticipant removes a participant from a thread.
func (s *ThreadsService) RemoveParticipant(threadID string, participantID string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", "/api/v1/threads/"+threadID+"/participants/"+participantID, nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateParticipant updates a participant's display name, avatar, or mute state.
func (s *ThreadsService) UpdateParticipant(threadID string, participantID string, input UpdateParticipantInput) (*ThreadParticipantInfo, error) {
	var result ThreadParticipantInfo
	err := s.client.doRequestTyped("PATCH", "/api/v1/threads/"+threadID+"/participants/"+participantID, input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Read States ─────────────────────────────────────────────────────────────

// GetReadStates returns read states for all participants in a thread.
func (s *ThreadsService) GetReadStates(threadID string) (*ReadStatesResponse, error) {
	var result ReadStatesResponse
	err := s.client.doRequestTyped("GET", "/api/v1/threads/"+threadID+"/read-state", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Cross-Thread Flags ──────────────────────────────────────────────────────

// ListFlagsAcrossThreads returns flags across all threads with optional filters.
func (s *ThreadsService) ListFlagsAcrossThreads(params map[string]string) (*FlagListResponse, error) {
	var result FlagListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/threads/flags", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ReviewFlag reviews a flagged message (approve/reject).
func (s *ThreadsService) ReviewFlag(flagID string, input ReviewFlagInput) (*ThreadFlag, error) {
	var result ThreadFlag
	err := s.client.doRequestTyped("PATCH", "/api/v1/threads/flags/"+flagID, input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
