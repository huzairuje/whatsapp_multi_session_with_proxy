package auth

import (
	"fmt"
	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/utils" // For utils.ParseJID

	"go.mau.fi/whatsmeow/types"
)

// SendOTPMessage sends an OTP to the given recipient phone number using a designated sender JID.
// cmdHandler: Instance of CommandHandler to send the message.
// otpSenderJIDString: The JID of the WhatsApp account that will send the OTP message (e.g., "1234567890@s.whatsapp.net").
// recipientPhoneNumber: The phone number of the admin user to send the OTP to (e.g., "0987654321").
// otp: The One-Time Password string to send.
func SendOTPMessage(cmdHandler *commandhandler.CommandHandler, otpSenderJIDString string, recipientPhoneNumber string, otp string) error {
	if cmdHandler == nil {
		return fmt.Errorf("commandHandler cannot be nil")
	}
	if otpSenderJIDString == "" {
		return fmt.Errorf("otpSenderJIDString cannot be empty")
	}
	if recipientPhoneNumber == "" {
		return fmt.Errorf("recipientPhoneNumber cannot be empty")
	}
	if otp == "" {
		return fmt.Errorf("otp cannot be empty")
	}

	// Parse the sender's JID string into a types.JID struct.
	// utils.ParseJID can handle JIDs like "number@s.whatsapp.net"
	// or just "number" and append "@s.whatsapp.net" if it's a valid WhatsApp user JID.
	senderJID, ok := utils.ParseJID(otpSenderJIDString)
	if !ok {
		return fmt.Errorf("failed to parse otpSenderJIDString '%s' into a valid JID", otpSenderJIDString)
	}

	// Construct the message content.
	messageContent := fmt.Sprintf("Your verification code for the dashboard is: %s", otp)

	// Call the CommandHandler's method to send the text message.
	// HandleSendTextMessage expects the sender JID (types.JID), the message content (string),
	// and the recipient's phone number (string), which it will internally parse into a JID.
	// The first return value of HandleSendTextMessage is a timestamp string, which we can ignore here.
	_, err := cmdHandler.HandleSendTextMessage(senderJID, messageContent, recipientPhoneNumber)
	if err != nil {
		// Wrap the error from HandleSendTextMessage for more context.
		return fmt.Errorf("failed to send OTP message via CommandHandler: %w", err)
	}

	return nil
}
