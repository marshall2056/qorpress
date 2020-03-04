package httpcache

// NopCache provides a no-op cache implementation that doesn't actually cache anything.
// var NotFound = new(notFound)

type NotFound struct{}

func (c *NotFound) Get(string) ([]byte, bool) { return nil, false }
func (c *NotFound) Set(string, []byte)        {}
func (c *NotFound) Delete(string)             {}
