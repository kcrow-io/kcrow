package errors

type K8sError struct{}

func (s *K8sError) Error() string {
	return "apiserver not response"
}

func (s *K8sError) Is(e error) bool {
	switch e.(type) {
	case *K8sError:
		return true
	default:
		return false
	}
}

type InternalError struct{}

func (s *InternalError) Error() string {
	return "internal error"
}

func (s *InternalError) Is(e error) bool {
	switch e.(type) {
	case *InternalError:
		return true
	default:
		return false
	}
}
