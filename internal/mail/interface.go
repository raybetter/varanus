package mail

type MailMessage struct {
	Sender    string
	Recipient string
	Subject   string
	Body      string
}

type MailWorker interface {
	SendMessage(accountName string, message MailMessage) error
	ReadMessage(accountName string, expectedSubject string) (MailMessage, error)
}
