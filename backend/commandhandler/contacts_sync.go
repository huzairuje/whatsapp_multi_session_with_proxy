package commandhandler

import (
	"context"
	"whatsapp_multi_session/contacts"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

func (ch CommandHandler) SyncContactsFromWhatsApp(sender types.JID, force bool) {
	client, ok := LoadClientConcurrent(sender.User)
	if !ok || client == nil {
		log.Errorf("[Contacts] Client not found for sender %s", sender.User)
		return
	}

	if !client.IsLoggedIn() {
		log.Errorf("[Contacts] Client not logged in for sender %s", sender.User)
		return
	}

	log.Infof("[Contacts] Starting contact sync for sender %s", sender.User)

	if ch.ContactsService == nil {
		log.Errorf("[Contacts] ContactsService not initialized for sender %s", sender.User)
		return
	}

	syncedCount := 0
	senderJID := sender.String()

	if client.Store == nil || client.Store.Contacts == nil {
		log.Warnf("[Contacts] No contacts store available for sender %s", sender.User)
		log.Infof("[Contacts] Sync completed for sender %s: %d contacts synced", sender.User, syncedCount)
		return
	}

	ctx := context.Background()
	contactMap, err := client.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		log.Errorf("[Contacts] Failed to get all contacts for sender %s: %v", sender.User, err)
		return
	}

	for jid, contact := range contactMap {
		if !contact.Found {
			continue
		}

		contactModel := contacts.Contact{
			SenderJID:    senderJID,
			ContactJID:   jid.String(),
			ContactName:  contact.FullName,
			PushName:     contact.PushName,
			BusinessName: contact.BusinessName,
			FirstName:    contact.FirstName,
			FullName:     contact.FullName,
			IsBlocked:    false,
			IsBusiness:   contact.BusinessName != "",
			IsEnterprise: false,
		}

		if err := ch.ContactsService.UpsertContact(contactModel); err != nil {
			log.Warnf("[Contacts] Failed to upsert contact %s for sender %s: %v", jid.String(), sender.User, err)
			continue
		}

		syncedCount++
	}

	log.Infof("[Contacts] Sync completed for sender %s: %d contacts synced", sender.User, syncedCount)
}
