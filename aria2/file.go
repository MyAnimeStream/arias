package aria2

type File struct {
	Index           int `json:",string"`
	Path            string
	Length          uint `json:",string"`
	CompletedLength uint `json:",string"`
	Selected        bool `json:",string"`
	URIs            []URI
}
