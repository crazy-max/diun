package model

// App holds application details
type App struct {
	ID      string
	Name    string
	Desc    string
	URL     string
	Author  string
	Version string
}

const (
	ImageStatusNew      = ImageStatus("new")
	ImageStatusUpdate   = ImageStatus("update")
	ImageStatusUnchange = ImageStatus("unchange")
)

// ImageStatus holds Docker image status analysis
type ImageStatus string
