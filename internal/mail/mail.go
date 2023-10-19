package mail

import (
	"fmt"
	"io"
	"strings"
	"time"
	"varanus/internal/config"
	"varanus/internal/secrets"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/kr/pretty"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func MakeMailWorker(config config.MailConfig, unsealer secrets.SecretUnsealer) MailWorker {
	return &mailWorkerImpl{
		config:   config,
		unsealer: unsealer,
	}
}

type mailWorkerImpl struct {
	config   config.MailConfig
	unsealer secrets.SecretUnsealer
}

// CanSend takes the account name and returns True if a message can be sent from the account,
// or false if not.  If false is returned, the bool return value contains the wait time.
func (mw *mailWorkerImpl) CanSend(accountName string) (bool, time.Duration) {
	//TODO implement the send limits here
	return true, time.Duration(0)
}

func (mw *mailWorkerImpl) SendMessage(accountName string, message MailMessage) error {
	//validate the message
	if len(message.Sender) > 0 {
		return fmt.Errorf("invalid message with non-empty Sender field.  Sender is set by the account")
	}
	if len(message.Recipient) == 0 {
		return fmt.Errorf("invalid message with empty Recipient field.  Recipient is required")
	}
	if len(message.Subject) == 0 {
		return fmt.Errorf("invalid message with empty Subject field.  Subject is required")
	}
	if len(message.Body) == 0 {
		return fmt.Errorf("invalid message with empty Body field.  Body is required")
	}

	account := mw.config.GetAccountByName(accountName)
	if account == nil {
		return fmt.Errorf("no account named '%s' was found", accountName)
	}
	if account.SMTP == nil {
		return fmt.Errorf("the account named '%s' has no SMTP config", accountName)
	}

	canSend, sendWait := mw.CanSend(accountName)
	if !canSend {
		log.Trace().Dur("sendWait", sendWait).Msg("Must wait to send message")
		return WaitError{sendWait}
	}

	log.Trace().Str("accountName", accountName).Msg("Ready to send a message")

	unsealedPassword, err := account.SMTP.Password.ReadSecret(mw.unsealer)
	if err != nil {
		log.Trace().Err(err).Msg("Failed to unseal password secret")
		return fmt.Errorf("failed to unseal password secret: %w", err)
	}

	auth := sasl.NewPlainClient("", account.SMTP.Username, unsealedPassword)

	msgBody := fmt.Sprintf("To: %s\r\n", message.Recipient) +
		fmt.Sprintf("Subject: %s\r\n", message.Subject) +
		"\r\n" +
		fmt.Sprintf("%s\r\n", message.Body)

	mailServerAddress := fmt.Sprintf("%s:%d", account.SMTP.ServerAddress, account.SMTP.Port)

	log.Trace().Str("mailServerAddress", mailServerAddress).Msg("Sending to")

	// Connect to the remote SMTP server.
	var smtpClient *smtp.Client
	{
		var err error
		if account.SMTP.UseTLS {
			smtpClient, err = smtp.DialTLS(mailServerAddress, nil)
		} else {
			smtpClient, err = smtp.Dial(mailServerAddress)
		}
		if err != nil {
			log.Trace().Err(err).Str("mailServerAddress", mailServerAddress).Msgf("Failed to dial SMTP server")
			return fmt.Errorf("failed to dial SMTP server '%s': %w", mailServerAddress, err)
		}
	}
	//authenticate
	if err := smtpClient.Auth(auth); err != nil {
		//no test coverage for failures that require inducing an error in the SMTP server
		log.Trace().Err(err).Msgf("Failed to authenticate")
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set the sender and recipient first
	if err := smtpClient.Mail(account.SMTP.SenderAddress, nil); err != nil {
		//no test coverage for failures that require inducing an error in the SMTP server
		log.Trace().Err(err).Str("senderAddress", account.SMTP.SenderAddress).Msgf("Failed to set sender address")
		return fmt.Errorf("failed to set sender address '%s': %w", account.SMTP.SenderAddress, err)
	}
	if err := smtpClient.Rcpt(message.Recipient, nil); err != nil {
		//no test coverage for failures that require inducing an error in the SMTP server
		log.Trace().Err(err).Str("recipientAddress", message.Recipient).Msgf("Failed to set recipient address")
		return fmt.Errorf("failed to set recipient address '%s': %w", message.Recipient, err)
	}
	// Send the email body.
	{
		wc, err := smtpClient.Data()
		if err != nil {
			//no test coverage for failures that require inducing an error in the SMTP server
			log.Trace().Err(err).Str("email body", message.Body).Msgf("Failed to set email body")
			return fmt.Errorf("failed to set email body '%s': %w", message.Body, err)
		}
		if _, err = fmt.Fprint(wc, msgBody); err != nil {
			//no test coverage for failures that require inducing an error in the SMTP server
			log.Trace().Err(err).Str("email body", message.Body).Msgf("Failed to write body to message")
			return fmt.Errorf("failed to write body to message '%s': %w", message.Body, err)
		}
		if err := wc.Close(); err != nil {
			//no test coverage for failures that require inducing an error in the SMTP server
			log.Trace().Err(err).Msgf("Failed to close email body")
			return fmt.Errorf("failed to close email body: %w", err)
		}
	}

	// Send the QUIT command and close the connection.
	if err := smtpClient.Quit(); err != nil {
		//no test coverage for failures that require inducing an error in the SMTP server
		log.Trace().Err(err).Msgf("Failed to quit client")
		return fmt.Errorf("failed to quit client: %w", err)
	}

	//success
	return nil
}

func (mw *mailWorkerImpl) ReadMessage(accountName string, expectedSubject string) (MailMessage, error) {
	//get the account
	account := mw.config.GetAccountByName(accountName)
	if account == nil {
		return MailMessage{}, fmt.Errorf("no account named '%s' was found", accountName)
	}
	if account.IMAP == nil {
		return MailMessage{}, fmt.Errorf("the account named '%s' has no IMAP config", accountName)
	}

	unsealedPassword, err := account.IMAP.Password.ReadSecret(mw.unsealer)
	if err != nil {
		return MailMessage{}, fmt.Errorf("failed to unseal password secret: %w", err)
	}

	// Connect to server
	mailServerAddress := fmt.Sprintf("%s:%d", account.IMAP.ServerAddress, account.IMAP.Port)

	var imapClient *client.Client
	{
		var err error
		if account.IMAP.UseTLS {
			imapClient, err = client.DialTLS(mailServerAddress, nil)
		} else {
			imapClient, err = client.Dial(mailServerAddress)
		}
		if err != nil {
			return MailMessage{}, fmt.Errorf("failed to dial IMAP server %s: %w", mailServerAddress, err)
		}
	}

	// Don't forget to logout
	defer imapClient.Logout()

	// Login
	if err := imapClient.Login(account.IMAP.Username, unsealedPassword); err != nil {
		return MailMessage{}, fmt.Errorf("failed to login to IMAP server %s: %w",
			mailServerAddress, err)
	}

	mbox, err := imapClient.Select(account.IMAP.MailboxName, true)
	if err != nil {
		return MailMessage{}, fmt.Errorf("failed to select the mailbox %s: %w",
			account.IMAP.MailboxName, err)
	}

	const CHUNK_SIZE = 5
	chunks := makeReverseChunks(1, int(mbox.Messages), CHUNK_SIZE)

	var foundMessage *MailMessage
	messageCount := 0
	for _, chunk := range chunks {
		//range over messages in the chunk in ascending order
		seqset := new(imap.SeqSet)
		seqset.AddRange(uint32(chunk.start), uint32(chunk.end))

		messages := make(chan *imap.Message, CHUNK_SIZE)
		done := make(chan error, 1)

		//call a go-routine to fetch messages
		go func() {
			done <- imapClient.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		//read from the messages channel until Fetch closes it
		for msg := range messages {
			messageCount += 1
			if msg.Envelope.Subject == expectedSubject {
				//found the message we are looking for, so fetch the message body
				bodyText, err := getBodyText(imapClient, msg.SeqNum)
				if err != nil {
					bodyText = "Unable to get body text."
					log.Warn().Err(err).Interface("envelope msg", msg).Msg("Unable to fetch the message body")
				}

				foundMessage = &MailMessage{
					Subject:   msg.Envelope.Subject,
					Recipient: addressesToString(msg.Envelope.To),
					Sender:    addressesToString(msg.Envelope.Sender),
					Body:      bodyText,
				}
				break
			}
		}
		if foundMessage != nil {
			break
		}

		if err := <-done; err != nil {
			return MailMessage{}, fmt.Errorf("failed while fetching messages: %w", err)
		}

		if messageCount > 10 {
			break
		}
	}

	if foundMessage == nil {
		return MailMessage{}, fmt.Errorf("no message matching the subject '%s' was found", expectedSubject)
	}

	return *foundMessage, nil
}

func getBodyText(client *client.Client, messageSeqNum uint32) (string, error) {
	//sequence set for the message we are targeting
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(messageSeqNum)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- client.Fetch(seqSet, items, messages)
	}()

	msg := <-messages
	if msg == nil {
		log.Trace().Msg("While fetching message body, server fetch didn't return anything")
		return "", fmt.Errorf("while fetching message body, server fetch didn't return anything")
	}

	log.Trace().Msg("**********************************************")
	log.Trace().Msgf("%# v", pretty.Formatter(msg))
	log.Trace().Msg("**********************************************")

	if err := <-done; err != nil {
		log.Trace().Err(err).Msg("While fetching message body")
		return "", fmt.Errorf("error fetching message body: %w", err)
	}

	r := msg.GetBody(&section)
	if r == nil {
		log.Trace().Interface("message", msg).Msg("no body found in returned message")
		return "", fmt.Errorf("no body found in returned message")
	}

	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Trace().Err(err).Msg("failed to create mail reader")
		return "", fmt.Errorf("failed to create mail reader: %w", err)
	}

	// Process each message's part
	var sb strings.Builder
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Trace().Err(err).Msg("unexpected error while reading parts of the message")
			return "", fmt.Errorf("unexpected error while reading parts of the message: %w", err)
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			// This is the message's text (can be plain-text or HTML)
			b, err := io.ReadAll(p.Body)
			if err != nil {
				log.Trace().Err(err).Interface("part", p).Msg("unexpected error while reading the message part")
				return "", fmt.Errorf("unexpected error while reading the message part: %w", err)
			}
			sb.Write(b)
		case *mail.AttachmentHeader:
			// This is an attachment
			filename, _ := h.Filename()
			log.Trace().Str("filename", filename).Msg("Skipping attachment")
		default:
			log.Trace().Interface("part", p).Msg("Skipping other part")
		}
	}

	return strings.TrimSpace(sb.String()), nil

}
