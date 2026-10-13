package service

// ImageCapabilities reports what the compiled image backend can do. Slim
// builds (`-tags novips`, pure Go) cannot encode WebP or render watermarks.
type ImageCapabilities struct {
	Backend    string `json:"backend"`
	WebPOutput bool   `json:"webp_output"`
	Watermark  bool   `json:"watermark"`
}
