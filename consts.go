package uploader

// transport constants
const (
	BlockSize = 1024 * 1024
	FilePerm  = 0777

	NetTCP = "tcp"

	ServicePath = "RPC"

	MethodCreateFile = "Create"
	MethodCreatePath = "CreatePath"
	MethodStat       = "Stat"
	WriteAt          = "WriteAt"
	Close            = "Close"
)

const (
	CREATE = "CREATE"
)

const (
	TypeDirectory = "Directory"
	TypeFile      = "File"
)
