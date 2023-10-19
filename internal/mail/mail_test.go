package mail

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
	"varanus/internal/config"
	"varanus/internal/secrets"
	"varanus/internal/util"
	"varanus/internal/validation"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailWorker(t *testing.T) {

	//this test does a round trip of sending and checking for a message with a real mailserver.
	//to protect the confidentiality of the account password, the sealed config file and password
	//protected private key are part of the repo, but the passphrase must be set in VARANUS_REMOTE_MAIL_TEST_PASSPHRASE before the test is run

	util.SlowTest(t)

	key_passphrase := os.Getenv("VARANUS_REMOTE_MAIL_TEST_PASSPHRASE")
	if key_passphrase == "" {
		t.Skip("Skipping TestMailWorker in internal/mail/mail_test.go because VARANUS_REMOTE_MAIL_TEST_PASSPHRASE is not set")
		return
	}

	testConfig, err := config.ReadConfigFromFile("tests/314pies_test_config.sealed.yaml")
	require.Nil(t, err)

	{
		//we should always test with a valid config
		result, err := validation.ValidateObject(testConfig)
		require.Nil(t, err)
		require.Equal(t, 0, result.GetErrorCount())
	}

	unsealer := secrets.MakeSecretUnsealer()
	unsealer.LoadPrivateKeyFromFile("tests/314pies_test_key.pem", key_passphrase)
	checkResult := unsealer.CheckSeals(testConfig)
	require.Len(t, checkResult.UnsealErrors, 0)

	//if we get here, the config is loaded and valid, so now we can test

	subjectLine := "test message " + time.Now().Format(time.DateTime)

	worker := mailWorkerImpl{testConfig.Mail, unsealer}

	err = worker.SendMessage("314pies_account", MailMessage{
		Recipient: "mailtest2@314pies.com",
		Subject:   subjectLine,
		Body:      "This is the message body.",
	})
	require.Nil(t, err)

	time.Sleep(time.Duration(10) * time.Second)

	message, err := worker.ReadMessage("314pies_account", subjectLine)
	assert.Nil(t, err)
	assert.Equal(t, "mailtest2@314pies.com", message.Recipient)
	assert.Equal(t, subjectLine, message.Subject)
	assert.Equal(t, "This is the message body.", message.Body)

}

func TestMailWorkerLocal(t *testing.T) {

	testRun := func() error {
		fmt.Fprintln(os.Stderr, "Starting test operation")
		// email addresses, usernames, and passwords don't matter for this test because the test
		// mailserver will accept anything
		config := config.VaranusConfig{
			Mail: config.MailConfig{
				Accounts: []config.MailAccountConfig{
					{
						Name: "account1",
						SMTP: &config.SMTPConfig{
							SenderAddress: "mailtest2@314pies.com",
							ServerAddress: "localhost",
							Port:          2525,
							UseTLS:        false,
							Username:      "mailtest2@314pies.com",
							Password:      secrets.CreateSealedItem("random password"),
						},
						IMAP: &config.IMAPConfig{
							RecipientAddress: "mailtest2@314pies.com",
							MailboxName:      "INBOX",
							ServerAddress:    "localhost",
							Port:             2543,
							UseTLS:           false,
							Username:         "mailtest2@314pies.com",
							Password:         secrets.CreateSealedItem("random password"),
						},
					},
				},
			},
		}

		{
			//we should always test with a valid config
			result, err := validation.ValidateObject(config)
			require.Nil(t, err)
			require.Equal(t, 0, result.GetErrorCount())
		}

		worker := MakeMailWorker(config.Mail, nil)

		fmt.Fprintln(os.Stderr, "Sending mail")

		subjectLine := "test message 1" + time.Now().Format(time.DateTime)

		err := worker.SendMessage("account1", MailMessage{
			Recipient: "mailtest2@314pies.com",
			Subject:   subjectLine,
			Body:      "This is the message body.",
		})
		assert.Nil(t, err)

		fmt.Fprintln(os.Stderr, "Done sending mail")

		time.Sleep(500 * time.Millisecond)

		fmt.Fprintln(os.Stderr, "Checking mail")

		message, err := worker.ReadMessage("account1", subjectLine)
		assert.Nil(t, err)
		assert.Equal(t, "mailtest2@314pies.com", message.Recipient)
		assert.Equal(t, subjectLine, message.Subject)
		assert.Equal(t, "This is the message body.", message.Body)

		fmt.Fprintln(os.Stderr, "Done checking mail")

		fmt.Fprintln(os.Stderr, "Test operation completed")

		return nil
	}

	//execute the test function with a local mail server docker container running
	//equivalent to: docker run --rm -it -p 3000:80 -p 2525:25 -p 2543:143 rnwood/smtp4dev

	readyFunc := func() bool {
		fmt.Fprintln(os.Stderr, "checking for mail server readiness")
		resp, err := http.Get("http://localhost:3000/")
		if err != nil {
			fmt.Fprintln(os.Stderr, "in ready function, http error was", err)
			return false
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintln(os.Stderr, "in ready function, status code was", resp.StatusCode)
			return false
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "in ready function, failed to read body", err)
		}
		fmt.Fprintln(os.Stderr, "Ready function succeded.  The returned body was", string(bodyBytes))

		return true
	}

	operation := util.DockerContextOperation{
		MaxReadyWait:      1200 * time.Second,
		ReadyWait:         1 * time.Second,
		ReadyCallback:     readyFunc,
		ExecutionCallback: testRun,
	}

	options := dockertest.RunOptions{
		Repository: "rnwood/smtp4dev",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"80/tcp":  {{HostIP: "localhost", HostPort: "3000/tcp"}},
			"25/tcp":  {{HostIP: "localhost", HostPort: "2525/tcp"}},
			"143/tcp": {{HostIP: "localhost", HostPort: "2543/tcp"}},
		},
	}

	err := util.TheDockerContext.ExecuteWithOptions(
		&options,
		&operation,
	)
	assert.Nil(t, err)

}

func TestMailWorkerInvalidSend(t *testing.T) {
	config := config.VaranusConfig{
		Mail: config.MailConfig{
			Accounts: []config.MailAccountConfig{
				{
					Name: "account1",
					SMTP: &config.SMTPConfig{
						SenderAddress: "mailtest2@314pies.com",
						ServerAddress: "localhost",
						Port:          2525,
						UseTLS:        false,
						Username:      "mailtest2@314pies.com",
						Password:      secrets.CreateSealedItem("some password"),
					},
					IMAP: &config.IMAPConfig{
						RecipientAddress: "mailtest2@314pies.com",
						MailboxName:      "INBOX",
						ServerAddress:    "localhost",
						Port:             2543,
						UseTLS:           false,
						Username:         "mailtest2@314pies.com",
						Password:         secrets.CreateSealedItem("some password"),
					},
				},
				{
					//has no SMTP config
					Name: "account2",
					IMAP: &config.IMAPConfig{
						RecipientAddress: "mailtest2@314pies.com",
						MailboxName:      "INBOX",
						ServerAddress:    "localhost",
						Port:             2543,
						UseTLS:           false,
						Username:         "mailtest2@314pies.com",
						Password:         secrets.CreateSealedItem("some password"),
					},
				},
				{
					//has an invalid sealed password
					Name: "account3",
					SMTP: &config.SMTPConfig{
						SenderAddress: "mailtest2@314pies.com",
						ServerAddress: "localhost",
						Port:          2525,
						UseTLS:        false,
						Username:      "mailtest2@314pies.com",
						Password:      secrets.CreateSealedItem("sealed(+aaaaaa==)"),
					},
				},
			},
		},
	}

	{
		//we should always test with a valid config
		result, err := validation.ValidateObject(config)
		require.Nil(t, err)
		require.Equal(t, 0, result.GetErrorCount())
	}

	worker := MakeMailWorker(config.Mail, nil)

	{
		err := worker.SendMessage("account1", MailMessage{
			Sender: "should be empty",
		})
		assert.ErrorContains(t, err, "non-empty Sender field")
	}
	{
		err := worker.SendMessage("account1", MailMessage{
			Subject: "test subject",
			Body:    "This is the message body.",
		})
		assert.ErrorContains(t, err, "empty Recipient field")
	}
	{
		err := worker.SendMessage("account1", MailMessage{
			Recipient: "mailtest2@314pies.com",
			Body:      "This is the message body.",
		})
		assert.ErrorContains(t, err, "empty Subject field")
	}
	{
		err := worker.SendMessage("account1", MailMessage{
			Recipient: "mailtest2@314pies.com",
			Subject:   "test subject",
		})
		assert.ErrorContains(t, err, "empty Body field")
	}
	{
		err := worker.SendMessage("nonexistent_account", MailMessage{
			Recipient: "mailtest2@314pies.com",
			Subject:   "test subject",
			Body:      "This is the message body.",
		})
		assert.ErrorContains(t, err, "no account named 'nonexistent_account' was found")
	}
	{
		err := worker.SendMessage("account2", MailMessage{
			Recipient: "mailtest2@314pies.com",
			Subject:   "test subject",
			Body:      "This is the message body.",
		})
		assert.ErrorContains(t, err, "has no SMTP config")
	}
	{
		err := worker.SendMessage("account2", MailMessage{
			Recipient: "mailtest2@314pies.com",
			Subject:   "test subject",
			Body:      "This is the message body.",
		})
		assert.ErrorContains(t, err, "has no SMTP config")
	}
	{
		err := worker.SendMessage("account3", MailMessage{
			Recipient: "mailtest2@314pies.com",
			Subject:   "test subject",
			Body:      "This is the message body.",
		})
		assert.ErrorContains(t, err, "failed to unseal password secret")
	}
	{
		//everything should be valid, but the SMPT server is not running
		err := worker.SendMessage("account1", MailMessage{
			Recipient: "mailtest2@314pies.com",
			Subject:   "test subject",
			Body:      "This is the message body.",
		})
		assert.ErrorContains(t, err, "failed to dial SMTP server")
	}

}
