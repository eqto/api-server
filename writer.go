package api

type Writer interface {
	//Write data, return length has been written and errors if any
	Write(data []byte) (int, error)
	Flush() error
	Close() error
}
