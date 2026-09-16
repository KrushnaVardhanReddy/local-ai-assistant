package driven

type AudioCapturePort interface {
	Initialize() error
	StartCapture(deviceIndex int, includeLoopback bool, callback func([]float32)) error
	Stop()
}
