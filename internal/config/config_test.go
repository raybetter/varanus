package config

import (
	"encoding/json"
	"reflect"
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
	assert.Equal(t, "+aaaaaa==", c.Mail.Accounts[0].SMTP.Password.SealedValue)

	assert.NotNil(t, c.Mail.Accounts[0].IMAP)
	assert.Equal(t, "imap.example.com", c.Mail.Accounts[0].IMAP.ServerAddress)
	assert.Equal(t, uint(993), c.Mail.Accounts[0].IMAP.Port)
	assert.Equal(t, "janeuser@example.com", c.Mail.Accounts[0].IMAP.Username)
	assert.Equal(t, "+bbbbbb==", c.Mail.Accounts[0].IMAP.Password.SealedValue)

	assert.Len(t, c.Mail.SendLimits, 1)
	assert.Equal(t, c.Mail.SendLimits[0].MinPeriodMinutes, 10)
	assert.Equal(t, c.Mail.SendLimits[0].AccountNames, []string{"test1"})

}

func TestInvalidFilename(t *testing.T) {
	filename := "tests/nonexistent.yaml"

	c, err := ReadConfig(filename)

	assert.Nil(t, c)
	// fmt.Printf("Error: %s", err)
	assert.ErrorContains(t, err, "file read error for config file")

}

func TestInvalidYaml(t *testing.T) {
	filename := "tests/invalid.yaml"

	c, err := ReadConfig(filename)

	assert.Nil(t, c)
	// fmt.Printf("Error: %s", err)
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
        password: sealed(+aaa/aaa==)
      imap:
        server_address: "imap.example.com"
        port: 993
        username: janeuser@example.com
        password: sealed(+bbb/bbb==)
  send_limits: []
            `,
			Result: VaranusConfig{
				Mail: MailConfig{
					Accounts: []MailAccountConfig{
						{
							Name: "test1",
							SMTP: &SMTPConfig{
								SenderAddress: "example@example.com",
								ServerAddress: "smtp.example.com",
								Port:          uint(465),
								Username:      "joeuser@example.com",
								Password:      SealedItem{"+aaa/aaa=="},
							},
							IMAP: &IMAPConfig{
								ServerAddress: "imap.example.com",
								Port:          uint(993),
								Username:      "janeuser@example.com",
								Password:      SealedItem{"+bbb/bbb=="},
							},
						},
					},
					SendLimits: []SendLimit{},
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
        password: sealed(+abcdef==)
  send_limits: []
            `,
			Result: VaranusConfig{
				Mail: MailConfig{
					Accounts: []MailAccountConfig{
						{
							Name: "test1",
							SMTP: &SMTPConfig{
								SenderAddress: "example@example.com",
								ServerAddress: "smtp.example.com",
								Port:          uint(465),
								Username:      "joeuser@example.com",
								Password:      SealedItem{"+abcdef=="},
							},
							IMAP: nil,
						},
					},
					SendLimits: []SendLimit{},
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
        password: sealed(+abcdef==)
  send_limits: []
            `,
			Result: VaranusConfig{
				Mail: MailConfig{
					Accounts: []MailAccountConfig{
						{
							Name: "test1",
							SMTP: nil,
							IMAP: &IMAPConfig{
								ServerAddress: "imap.example.com",
								Port:          uint(993),
								Username:      "janeuser@example.com",
								Password:      SealedItem{"+abcdef=="},
							},
						},
					},
					SendLimits: []SendLimit{},
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

func TestVaranusConfigValidation(t *testing.T) {

	type TestCase struct {
		Mutator         func(c *VaranusConfig)
		Error           string
		ErrorObjectType interface{}
	}

	testCases := []TestCase{
		//pass one error through to the MailConfig structs to check end to end behavior
		{
			Mutator:         func(c *VaranusConfig) { c.Mail.Accounts[0].SMTP.Username = "" },
			Error:           "username must not be empty or whitespace",
			ErrorObjectType: SMTPConfig{},
		},
	}

	baseConfig := VaranusConfig{
		Mail: MailConfig{
			Accounts: []MailAccountConfig{
				{
					Name: "test1",
					SMTP: &SMTPConfig{
						SenderAddress: "example@example.com",
						ServerAddress: "mail.example.com",
						Port:          465,
						Username:      "joe@example.com",
						Password:      SealedItem{"+abcdef=="},
					},
					IMAP: &IMAPConfig{
						ServerAddress: "mail.example.com",
						Port:          993,
						Username:      "joe@example.com",
						Password:      SealedItem{"+abcdef=="},
					},
				},
			},
			SendLimits: []SendLimit{
				{
					MinPeriodMinutes: 15,
					AccountNames:     []string{"test1"},
				},
			},
		},
	}

	//nominal case test should have no errors
	vp := &ValidationProcess{}
	config := baseConfig
	//validation
	config.Validate(vp)
	//checks
	assert.Len(t, vp.ErrorList, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &ValidationProcess{}
		config := baseConfig
		testCase.Mutator(&config)
		//validation
		config.Validate(vp)
		//checks
		assert.Len(t, vp.ErrorList, 1, "for test %d", index)
		assert.IsType(t, testCase.ErrorObjectType, vp.ErrorList[0].Object, "for test %d", index)
		assert.Contains(t, vp.ErrorList[0].Error, testCase.Error, "for test %d", index)
	}

}

func assertContainsErrorText(t *testing.T, vpe ValidationError, errorText string) {
	for _, validationError := range vpe.ErrorList {
		if strings.Contains(validationError.Error, errorText) {
			return
		}
	}
	//no match, so the assertion fails
	t.Errorf(
		"No validation error containing '%s' in ValidationProcessError with %d errors: %s",
		errorText, len(vpe.ErrorList), vpe.ErrorList)
}

// TEstValidationErrors keeps a few yaml validation error cases to test the end to end with
// validation errors.  Most validation testing coverage is provided by the structs at the config
// object levels
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
	}

	for _, testCase := range testCases {
		yamldata := []byte(testCase.Input)
		config, err := parseAndValidateConfig(yamldata)
		assert.Nilf(t, config, "Config not nil for invalid test case %#v", testCase)

		vpe, ok := err.(ValidationError)
		// vpe.Print()
		require.Truef(t, ok, "The returned error should be a ValidationProcessError, not %#v", err)
		assert.Len(t, vpe.ErrorList, len(testCase.ValidationErrors))
		for _, validationErrorText := range testCase.ValidationErrors {
			assertContainsErrorText(t, vpe, validationErrorText)
		}
	}

}

// deepCopyForTesting is an exceedingly lazy but compact way of deep copying the config structs
// with all their pointer objects also duplicated
//
// https://stackoverflow.com/questions/50269322/how-to-copy-struct-and-dereference-all-pointers
func deepCopyForTesting(v interface{}) interface{} {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	vptr := reflect.New(reflect.TypeOf(v))
	err = json.Unmarshal(data, vptr.Interface())
	if err != nil {
		panic(err)
	}
	return vptr.Elem().Interface()
}
