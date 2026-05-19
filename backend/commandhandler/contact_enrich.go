package commandhandler

import (
	"whatsapp_multi_session/contacts"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

func (ch CommandHandler) EnrichRecipientWithContactInfo(senderJID types.JID, recipientPhone string, variables map[string]string) map[string]string {
	if ch.ContactsService == nil {
		return variables
	}

	enrichedVars := make(map[string]string)
	for k, v := range variables {
		enrichedVars[k] = v
	}

	filter := contacts.ContactFilter{
		SenderJID:   senderJID.String(),
		SearchQuery: recipientPhone,
		Limit:       1,
		Offset:      0,
	}

	contactList, err := ch.ContactsService.SearchContacts(filter)
	if err != nil {
		log.Debugf("[ContactEnrich] Failed to search contacts for %s: %v", recipientPhone, err)
		return enrichedVars
	}

	if len(contactList) == 0 {
		log.Debugf("[ContactEnrich] No contact found for %s", recipientPhone)
		return enrichedVars
	}

	contact := contactList[0]

	if contact.ContactName != "" {
		enrichedVars["contact_name"] = contact.ContactName
	}
	if contact.FirstName != "" {
		enrichedVars["first_name"] = contact.FirstName
	}
	if contact.FullName != "" {
		enrichedVars["full_name"] = contact.FullName
	}
	if contact.PushName != "" {
		enrichedVars["push_name"] = contact.PushName
	}
	if contact.BusinessName != "" {
		enrichedVars["business_name"] = contact.BusinessName
	}

	log.Debugf("[ContactEnrich] Enriched recipient %s with contact data: %s", recipientPhone, contact.ContactName)

	return enrichedVars
}
