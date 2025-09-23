package ports

type Verifier interface {
	Verify(header string, body []byte) error
}
