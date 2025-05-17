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
