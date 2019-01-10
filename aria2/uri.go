package aria2

type URIStatus string

const (
	URIUsed    URIStatus = "used"
	URIWaiting           = "waiting"
)

type URI struct {
	URI    string
	Status URIStatus
}
