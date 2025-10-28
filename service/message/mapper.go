package message

import (
	"qubit/env/postgres/messages"
)

// ToDomain converts a postgres Message model to a domain Message
func ToDomain(message *messages.Message) *Message {
	if message == nil {
		return nil
	}

	return &Message{
		ID:          message.ID,
		PhoneNumber: message.PhoneNumber,
		Content:     message.Content,
		CreatedAt:   message.CreatedAt,
		MessageID:   message.MessageID,
		ProcessedAt: message.ProcessedAt,
	}
}

// ToPostgres converts a domain Message to a postgres Message model
func ToPostgres(domainMsg *Message) *messages.Message {
	if domainMsg == nil {
		return nil
	}

	return &messages.Message{
		ID:          domainMsg.ID,
		PhoneNumber: domainMsg.PhoneNumber,
		Content:     domainMsg.Content,
		CreatedAt:   domainMsg.CreatedAt,
		MessageID:   domainMsg.MessageID,
		ProcessedAt: domainMsg.ProcessedAt,
	}
}

// ToDomainSlice converts a slice of postgres Messages to domain Messages
func ToDomainSlice(dbMessages []*messages.Message) []*Message {
	if dbMessages == nil {
		return nil
	}

	domainMessages := make([]*Message, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		domainMessages = append(domainMessages, ToDomain(dbMsg))
	}

	return domainMessages
}
