package sender

func NewMatrixSenderWithBase(baseURL, roomID, accessToken string) Sender {
	return newMatrixSenderWithBase(baseURL, roomID, accessToken)
}
