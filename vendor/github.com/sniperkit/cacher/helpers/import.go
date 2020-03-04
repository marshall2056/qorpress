package helpers

// newThrottler creates throttler that gives out resources up to allowed.
// Resources must be returned by re-sending on the channel.
func NewThrottler(allowed int) chan struct{} {
	ch := make(chan struct{}, allowed)
	for i := 0; i < allowed; i++ {
		ch <- struct{}{}
	}
	return ch
}
