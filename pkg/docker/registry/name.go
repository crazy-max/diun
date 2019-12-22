package registry

// Name returns the full name representation of an image.
func (i Image) Name() string {
	return i.named.Name()
}
