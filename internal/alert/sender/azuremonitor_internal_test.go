package sender

// NewAzureMonitorSenderWithURL exposes the internal constructor for testing.
func NewAzureMonitorSenderWithURL(workspaceID, sharedKey, logType, url string) *azureMonitorSender {
	return newAzureMonitorSenderWithURL(workspaceID, sharedKey, logType, url)
}
