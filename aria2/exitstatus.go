package aria2

type ExitStatus uint8

const (
	Success ExitStatus = iota
	UnknownError
	Timeout
	ResourceNotFound
	ResourceNotFoundReached
	DownloadSpeedTooSlow
	NetworkProblem
	UnfinishedDownloads
	RemoteNoResume
	NotEnoughDiskSpace
	PieceLengthMismatch
	SameFileBeingDownloaded
	SameInfoHashBeingDownloaded
	FileAlreadyExists
	RenamingFailed
	CouldNotOpenExistingFile
	CouldNotCreateNewFile
	FileIOError
	CouldNotCreateDirectory
	NameResolutionFailed
	MetalinkParsingFailed
	FTPCommandFailed
	HTTPResponseHeaderBad
	TooManyRedirects
	HttpAuthorizationFailed
	BEncodedFileParseError
	TorrentFileCorrupt
	MagnetURIBad
	RemoteServerHandleRequestError
	JSONRPCParseError
	Reserved
	ChecksumValidationFailed
)
