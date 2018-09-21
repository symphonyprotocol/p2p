package tcp

type SecuredTCPService struct {
	*TCPService
}

func NewSecuredTCPService() *SecuredTCPService {
	service := &SecuredTCPService{
		&TCPService{

		},
	}

	return service
}


