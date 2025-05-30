package primitive

const (
	ShutDownEvent = "ShutDownEvent"
)

const (
	MediaImage    = "image"
	MediaVideo    = "video"
	MediaAudio    = "audio"
	MediaDocument = "document"
	MediaHistory  = "md-msg-hist"
	MediaAppState = "md-app-state"

	MediaLinkThumbnail = "thumbnail-link"

	StatusSend      = iota + 1
	StatusDelivered = StatusSend + 1
	StatusRead      = StatusDelivered + 1

	NotFound = "not found"
	Deleted  = "token was delete"

	ImageJPEG = "image/jpeg"
	ImageJPG  = "image/jpg"
	ImagePNG  = "image/png"
	ImageWEBP = "image/webp"
	ImageGIF  = "image/gif"
	ImageAVIF = "image/avif"
	ImageAPNG = "image/apng"
	ImageSVG  = "image/svg+xml"

	VideoMp4       = "video/mp4"
	VideoMpeg      = "video/mpeg"
	VideoOgg       = "video/ogg"
	VideoWebm      = "video/webm"
	VideoAvi       = "video/avi"
	VideoQuickTime = "video/quicktime"
	VideoWmv       = "video/x-ms-wmv"

	AudioMpeg = "audio/mpeg"
	AudioOgg  = "audio/ogg"
	AudioWav  = "audio/wav"
	AudioWebm = "audio/webm"
	AudioAac  = "audio/aac"
	AudioMp4  = "audio/mp4"
)

const (
	MessageInvalidUploadFile       = "there is error on parameter file or parameter file is failed to upload"
	MessageInvalidRecipient        = "there is error on parameter recipients or parameter recipients is invalid or not found"
	MessageFailedToCloseFile       = "failed to close file, please check your file path or file name"
	MessageFailedToDownloadFile    = "failed to download file or the file is invalid URL"
	MessageFailedToReadFileData    = "failed to read file data, please check your file path or file name"
	MessageFileNotFound            = "file not found, please check your file path or file name"
	MessageErrorRequestMultiPart   = "there is error when parsing multipart form or the request body is not valid multipart type"
	MessageInvalidPhoneNumber      = `your number recipient is not contains "+"`
	MessageInvalidJson             = "error decoding JSON request or invalid JSON format"
	MessageFailedToSend            = "failed to send, please login or reconnect"
	MessageSenderShouldBeFilled    = "sender should be filled"
	MessageInvalidJidRequest       = "invalid jid request"
	MessageJidNotFound             = "your request jid is not found"
	MessageSuccessSent             = "message successfully sent"
	MessageTriggeredAutoLogin      = "success trigger auto login"
	MessageTriggeredAutoDisconnect = "success trigger auto disconnect"
	MessageAlreadyLoggedIn         = "you are already logged in, please logout first before login again"
	MessageFailedToSendData        = "failed to send data"
)
