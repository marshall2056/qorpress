package kvfscache

type ErrFailedDelete struct {
	Key string
}

func (e *ErrFailedDelete) Error() string {
	return "Failed to delete key=" + e.Key
}

type ErrNotSupported struct {
	Protocol string
}

func (e *ErrNotSupported) Error() string {
	return "Protocol not supported:" + e.Protocol
}
