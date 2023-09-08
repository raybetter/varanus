package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEnd(t *testing.T) {
	filename := "tests/example.yaml"

	c, err := ReadConfig(filename)

	require.Nil(t, err)

	assert.Len(t, c.Mail.Accounts, 1)
	assert.Equal(t, "test1", c.Mail.Accounts[0].Name)
	assert.NotNil(t, c.Mail.Accounts[0].SMTP)
	assert.Equal(t, "example@example.com", c.Mail.Accounts[0].SMTP.SenderAddress)
	assert.Equal(t, "smtp.example.com", c.Mail.Accounts[0].SMTP.ServerAddress)
	assert.Equal(t, uint(465), c.Mail.Accounts[0].SMTP.Port)
	assert.Equal(t, "joeuser@example.com", c.Mail.Accounts[0].SMTP.Username)
	assert.Equal(t, "joepassword", c.Mail.Accounts[0].SMTP.Password)

	assert.NotNil(t, c.Mail.Accounts[0].IMAP)
	assert.Equal(t, "imap.example.com", c.Mail.Accounts[0].IMAP.ServerAddress)
	assert.Equal(t, uint(993), c.Mail.Accounts[0].IMAP.Port)
	assert.Equal(t, "janeuser@example.com", c.Mail.Accounts[0].IMAP.Username)
	assert.Equal(t, "janepassword", c.Mail.Accounts[0].IMAP.Password)

	assert.Len(t, c.Mail.SendLimits, 0)

}

func TestInvalidFilename(t *testing.T) {
	filename := "tests/nonexistent.yaml"

	c, err := ReadConfig(filename)

	assert.Nil(t, c)
	fmt.Printf("Error: %s", err)
	assert.ErrorContains(t, err, "file read error for config file")

}

func TestInvalidYaml(t *testing.T) {
	filename := "tests/invalid.yaml"

	c, err := ReadConfig(filename)

	assert.Nil(t, c)
	fmt.Printf("Error: %s", err)
	assert.ErrorContains(t, err, "unmarshal error")

}

func TestYamlValidCases(t *testing.T) {
	type TestCase struct {
		Input  string
		Result VaranusConfig
	}

	testCases := []TestCase{
		//------------------------------------------------------------------------------------------
		//both SMTP and IMAP in a config
		// use different ports than the defaults to ensure they are picked up
		{
			Input: `---
mail:
  accounts:
    - name: test1
      smtp:
        sender_address: "example@example.com"
        server_address: "smtp.example.com"	
        port: 465
        username: joeuser@example.com
        password: joepassword
      imap:
        server_address: "imap.example.com"
        port: 993
        username: janeuser@example.com
        password: janepassword
  send_limits: []
            `,
			Result: VaranusConfig{
				Mail: MailConfig{
					Accounts: []ServerConfig{
						{
							Name: "test1",
							SMTP: &SMTPConfig{
								SenderAddress: "example@example.com",
								ServerAddress: "smtp.example.com",
								Port:          uint(465),
								Username:      "joeuser@example.com",
								Password:      "joepassword",
							},
							IMAP: &IMAPConfig{
								ServerAddress: "imap.example.com",
								Port:          uint(993),
								Username:      "janeuser@example.com",
								Password:      "janepassword",
							},
						},
					},
					SendLimits: nil,
				},
			},
		},
		//------------------------------------------------------------------------------------------
		//------------------------------------------------------------------------------------------
		//SMTP but no IMAP config
		{
			Input: `---
mail:
  accounts:
    - name: test1
      smtp:
        sender_address: "example@example.com"
        server_address: "smtp.example.com"
        port: 465
        username: joeuser@example.com
        password: joepassword
  send_limits: []
            `,
			Result: VaranusConfig{
				Mail: MailConfig{
					Accounts: []ServerConfig{
						{
							Name: "test1",
							SMTP: &SMTPConfig{
								SenderAddress: "example@example.com",
								ServerAddress: "smtp.example.com",
								Port:          uint(465),
								Username:      "joeuser@example.com",
								Password:      "joepassword",
							},
							IMAP: nil,
						},
					},
					SendLimits: nil,
				},
			},
		},
		//------------------------------------------------------------------------------------------
		//------------------------------------------------------------------------------------------
		//IMAP but no SMTP in a config
		// use different ports than the defaults to ensure they are picked up
		{
			Input: `---
mail:
  accounts:
    - name: test1
      imap:
        server_address: "imap.example.com"
        port: 993
        username: janeuser@example.com
        password: janepassword
  send_limits: []
            `,
			Result: VaranusConfig{
				Mail: MailConfig{
					Accounts: []ServerConfig{
						{
							Name: "test1",
							SMTP: nil,
							IMAP: &IMAPConfig{
								ServerAddress: "imap.example.com",
								Port:          uint(993),
								Username:      "janeuser@example.com",
								Password:      "janepassword",
							},
						},
					},
					SendLimits: nil,
				},
			},
		},
		//------------------------------------------------------------------------------------------
	}

	for _, testCase := range testCases {
		yamldata := []byte(testCase.Input)
		config, err := parseAndValidateConfig(yamldata)
		assert.Nilf(t, err, "Error not nil for test case %#v", testCase)
		assert.Truef(t,
			assert.ObjectsAreEqualValues(&testCase.Result, config),
			"Configs not equal \nexpected: %# v\nactual: %# v",
			pretty.Formatter(testCase.Result),
			pretty.Formatter(config))
	}

}

func assertContainsErrorText(t *testing.T, vpe ValidationProcessError, errorText string) {
	for _, validationError := range vpe.Errors {
		if strings.Contains(validationError.Error, errorText) {
			return
		}
	}
	//no match, so the assertion fails
	t.Errorf("No validation error '%s' in ValidationProcessError with %d errors: %s", errorText, len(vpe.Errors), vpe.ErrorValue)
}

func TestValidationErrors(t *testing.T) {
	type TestCase struct {
		Input            string
		ValidationErrors []string
	}

	testCases := []TestCase{
		//------------------------------------------------------------------------------------------
		//Neither SMTP nor IMAP in a config
		{
			Input: `---
mail:
  accounts:
    - name: test1
  send_limits: []
`,
			ValidationErrors: []string{"Every server config must specify one of the imap or smtp sections"},
		},
		//------------------------------------------------------------------------------------------
		//------------------------------------------------------------------------------------------
		//SMTP validation
		{
			Input: `---
mail:
  accounts:
    - name: "  "
      smtp:
        sender_address: "foobar"
        server_address: "&&ThisisnotaURL&&&"
        port: 0
        username: "  "
        password: "  "
  send_limits: []
            `,
			ValidationErrors: []string{
				"account names must not be empty or whitespace",
				"sender_address 'foobar' is not a valid email",
				"is not a valid hostname",
				"port number is required",
				"username must not be empty or whitespace",
				"password must not be empty or whitespace",
			},
		},
		//------------------------------------------------------------------------------------------
		//------------------------------------------------------------------------------------------
		//IMAP validation
		{
			Input: `---
mail:
  accounts:
    - name: "  "
      imap:
        server_address: "&&ThisisnotaURL&&&"
        port: 0
        username: "  "
        password: "  "
  send_limits: []
            `,
			ValidationErrors: []string{
				"account names must not be empty or whitespace",
				"is not a valid hostname",
				"port number is required",
				"username must not be empty or whitespace",
				"password must not be empty or whitespace",
			},
		},
		//------------------------------------------------------------------------------------------
	}

	for _, testCase := range testCases {
		yamldata := []byte(testCase.Input)
		config, err := parseAndValidateConfig(yamldata)
		assert.Nilf(t, config, "Config not nil for invalid test case %#v", testCase)

		vpe, ok := err.(ValidationProcessError)
		vpe.Print()
		require.True(t, ok)
		assert.Len(t, vpe.Errors, len(testCase.ValidationErrors))
		for _, validationErrorText := range testCase.ValidationErrors {
			assertContainsErrorText(t, vpe, validationErrorText)
		}
	}

}
