package utils

import (
	"errors"
	"regexp"
	"strings"

	"whatsapp_multi_session/primitive"

	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow/types"
)

func StringContains(strSlice []string, str string) bool {
	for _, val := range strSlice {
		if val == str {
			return true
		}
	}
	return false
}

func IsImage(mimeType string) bool {
	extAsImage := []string{
		primitive.ImageJPEG,
		primitive.ImageJPG,
		primitive.ImagePNG,
		primitive.ImageWEBP,
		primitive.ImageGIF,
		primitive.ImageAVIF,
		primitive.ImageAPNG,
		primitive.ImageSVG,
	}
	return StringContains(extAsImage, mimeType)
}

func IsVideo(mimeType string) bool {
	extAsImage := []string{
		primitive.VideoMp4,
		primitive.VideoMpeg,
		primitive.VideoOgg,
		primitive.VideoWebm,
		primitive.VideoAvi,
		primitive.VideoQuickTime,
		primitive.VideoWmv,
	}
	return StringContains(extAsImage, mimeType)
}

func IsAudio(mimeType string) bool {
	extAsImage := []string{
		primitive.AudioMpeg,
		primitive.AudioOgg,
		primitive.AudioWav,
		primitive.AudioWebm,
		primitive.AudioAac,
		primitive.AudioMp4,
	}
	return StringContains(extAsImage, mimeType)
}

func ValidatePhoneNumber(phoneNumber string) bool {
	// Define a regular expression pattern
	pattern := `^\+[\d\s-]+$` // This pattern requires the phone number to start with '+'

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Check if the phone number matches the pattern
	return regex.MatchString(phoneNumber)

}

func GenerateQRCode(code string) ([]byte, error) {
	// Create QR code image
	qrImage, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	return qrImage, nil
}

// ParseJID Parse a JID from a string. If the string starts with a +, it is removed.
func ParseJID(arg string) (types.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			logrus.Errorf("Invalid JID %s: %v", arg, err)
			return recipient, false
		} else if recipient.User == "" {
			logrus.Errorf("Invalid JID %s: no server specified", arg)
			return recipient, false
		}
		return recipient, true
	}
}

func ValidateStringArrayAsStringArray(stringInput string) ([]string, error) {
	// Validate if the string is empty
	if strings.TrimSpace(stringInput) == "" {
		return nil, errors.New("empty string")
	}

	// Split the string by commas
	stringsArr := strings.Split(stringInput, ",")

	// Validate and clean each substring
	var stringSlice []string
	for _, str := range stringsArr {
		trimmedStr := strings.TrimSpace(str)
		// Append the converted value to the result slice
		stringSlice = append(stringSlice, trimmedStr)
	}

	return stringSlice, nil
}
