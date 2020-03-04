package limitedcache

// OpType is cache operation type enum.
type OpType uint8

const (
	// GetOp cache item operation.
	GetOp = OpType(iota)
	// SetOp cache item operation.
	SetOp
	// DeleteOp cache item operation.
	DeleteOp
)

// CacheOp - operations on cache with key and cache file name.
type CacheOp struct {
	op   OpType
	key  string
	file string
	err  error
}

// OperationID returns type of cache operation.
func (em *CacheOp) OperationID() OpType {
	return em.op
}

// Operation returns name of cache operation.
func (em *CacheOp) Operation() string {
	s := ""
	switch em.op {
	case GetOp:
		s = "get"
	case SetOp:
		s = "set"
	case DeleteOp:
		s = "delete"
	default:
		s = "unknown"
	}
	return s
}

// Key returns cache operation key.
func (em *CacheOp) Key() string {
	return em.key
}

// File returns cache operation filename on disk.
func (em *CacheOp) File() string {
	return em.file
}

// Status returns cache operation error.
func (em *CacheOp) Status() error {
	return em.err
}
