package config

import (
	"fmt"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	filename := "tests/example.yaml"

	c, err := ReadConfig(filename)

	assert.Nil(t, err)

	assert.Len(t, c.Mail.Accounts, 1)
	assert.Equal(t, "test1", c.Mail.Accounts[0].Name)
	assert.NotNil(t, c.Mail.Accounts[0].SMTP)
	assert.Equal(t, "example@example.com", c.Mail.Accounts[0].SMTP.SenderAddress)
	assert.Equal(t, "smtp.example.com", c.Mail.Accounts[0].SMTP.ServerAddress)
	assert.Equal(t, 465, c.Mail.Accounts[0].SMTP.Port)
	assert.Equal(t, "joeuser@example.com", c.Mail.Accounts[0].SMTP.Username)
	assert.Equal(t, "joepassword", c.Mail.Accounts[0].SMTP.Password)

	assert.NotNil(t, c.Mail.Accounts[0].IMAP)
	assert.Equal(t, "imap.example.com", c.Mail.Accounts[0].IMAP.ServerAddress)
	assert.Equal(t, 993, c.Mail.Accounts[0].IMAP.Port)
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
        port: 1465
        username: joeuser@example.com
        password: joepassword
      imap:
        server_address: "imap.example.com"
        port: 1993
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
								Port:          1465,
								Username:      "joeuser@example.com",
								Password:      "joepassword",
							},
							IMAP: &IMAPConfig{
								ServerAddress: "imap.example.com",
								Port:          1993,
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
								Port:          465,
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
        port: 1993
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
								Port:          1993,
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
		config, err := parseConfig(yamldata)
		assert.Nilf(t, err, "Error not nil for test case %#v", testCase)
		assert.Truef(t,
			assert.ObjectsAreEqualValues(&testCase.Result, config),
			"Configs not equal \nexpected: %# v\nactual: %# v",
			pretty.Formatter(testCase.Result),
			pretty.Formatter(config))
	}

}

// // TestHelloName calls greetings.Hello with a name, checking
// // for a valid return value.
// func TestHelloName(t *testing.T) {
//     name := "Gladys"
//     want := regexp.MustCompile(`\b`+name+`\b`)
//     msg, err := Hello("Gladys")
//     if !want.MatchString(msg) || err != nil {
//         t.Fatalf(`Hello("Gladys") = %q, %v, want match for %#q, nil`, msg, err, want)
//     }
// }

// // TestHelloEmpty calls greetings.Hello with an empty string,
// // checking for an error.
// func TestHelloEmpty(t *testing.T) {
//     msg, err := Hello("")
//     if msg != "" || err == nil {
//         t.Fatalf(`Hello("") = %q, %v, want "", error`, msg, err)
//     }
// }
