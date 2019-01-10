package aria2

type StatusName string

const (
	StatusActive    StatusName = "active"
	StatusWaiting              = "waiting"
	StatusPaused               = "paused"
	StatusError                = "error"
	StatusCompleted            = "completed"
	StatusRemoved              = "removed"
)

type Status struct {
	GID                    string
	Status                 StatusName
	TotalLength            uint `json:",string"`
	CompletedLength        uint `json:",string"`
	UploadLength           uint `json:",string"`
	BitField               string
	DownloadSpeed          uint `json:",string"`
	UploadSpeed            uint `json:",string"`
	InfoHash               string
	NumSeeders             uint       `json:",string"`
	Seeder                 bool       `json:",string"`
	PieceLength            uint       `json:",string"`
	NumPieces              uint       `json:",string"`
	Connections            uint       `json:",string"`
	ErrorCode              ExitStatus `json:",string"`
	ErrorMessage           string
	FollowedBy             []string
	Following              string
	BelongsTo              string
	Dir                    string
	Files                  []File
	Bittorrent             interface{}
	VerifiedLength         uint `json:",string"`
	VerifyIntegrityPending bool `json:",string"`
}
