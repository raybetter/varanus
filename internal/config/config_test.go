package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"varanus/internal/secrets"
	"varanus/internal/validation"

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
	assert.Equal(t, "it's a secret", c.Mail.Accounts[0].SMTP.Password.GetValue())

	assert.NotNil(t, c.Mail.Accounts[0].IMAP)
	assert.Equal(t, "imap.example.com", c.Mail.Accounts[0].IMAP.ServerAddress)
	assert.Equal(t, uint(993), c.Mail.Accounts[0].IMAP.Port)
	assert.Equal(t, "janeuser@example.com", c.Mail.Accounts[0].IMAP.Username)
	assert.Equal(t, "sealed(+bbbbbb==)", c.Mail.Accounts[0].IMAP.Password.GetValue())

	assert.Len(t, c.Mail.SendLimits, 1)
	assert.Equal(t, c.Mail.SendLimits[0].MinPeriodMinutes, 10)
	assert.Equal(t, c.Mail.SendLimits[0].AccountNames, []string{"test1"})

}

func TestEndToEndWithValidationErrors(t *testing.T) {
	filename := "tests/example-unvalidatable.yaml"

	_, err := ReadConfig(filename)

	assert.ErrorContains(t, err, "unable to parse config file")
	assert.ErrorContains(t, err, "account names must not be empty or whitespace")
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
			Input: `mail:
  accounts:
    - name: test1
      smtp:
        sender_address: example@example.com
        server_address: smtp.example.com
        port: 465
        username: joeuser@example.com
        password: sealed(+aaa/aaa==)
      imap:
        server_address: imap.example.com
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
								Password:      secrets.CreateSealedItem("sealed(+aaa/aaa==)"),
							},
							IMAP: &IMAPConfig{
								ServerAddress: "imap.example.com",
								Port:          uint(993),
								Username:      "janeuser@example.com",
								Password:      secrets.CreateSealedItem("sealed(+bbb/bbb==)"),
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
			Input: `mail:
  accounts:
    - name: test1
      smtp:
        sender_address: example@example.com
        server_address: smtp.example.com
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
								Password:      secrets.CreateSealedItem("sealed(+abcdef==)"),
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
			Input: `mail:
  accounts:
    - name: test1
      imap:
        server_address: imap.example.com
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
								Password:      secrets.CreateSealedItem("sealed(+abcdef==)"),
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
		//check the marshalling back to the string representation
		yamlString, err := config.ToYAML()
		assert.Nil(t, err)
		assert.Equal(t, testCase.Input, yamlString)
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
						Password:      secrets.CreateSealedItem("+abcdef=="),
					},
					IMAP: &IMAPConfig{
						ServerAddress: "mail.example.com",
						Port:          993,
						Username:      "joe@example.com",
						Password:      secrets.CreateSealedItem("+abcdef=="),
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
	vp := &validation.ValidationProcess{}
	config := baseConfig
	//validation
	config.Validate(vp)
	//checks
	assert.Len(t, vp.ErrorList, 0, "for nominal case")

	// test loop
	for index, testCase := range testCases {
		//setup
		vp := &validation.ValidationProcess{}
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

func assertContainsErrorText(t *testing.T, vpe validation.ValidationError, errorText string) {
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

		vpe, ok := err.(validation.ValidationError)
		// vpe.Print()
		require.Truef(t, ok, "The returned error should be a ValidationError, not %#v", err)
		assert.Len(t, vpe.ErrorList, len(testCase.ValidationErrors))
		for _, validationErrorText := range testCase.ValidationErrors {
			assertContainsErrorText(t, vpe, validationErrorText)
		}
	}

}

var testPublicKey string = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAybHNVmB0+9tNyZyKiMCJ
Qz0gUW/QtNdQd/XX26w6SCQtJVoM+o6r5vbz9YQoKHYe8vfeUdEmE79UorMMmndB
S8v8lMwWuCEy9MfcMRsSWnz8u9yRyhVfjaqEnMJHu2Pw05GLhQVf7fD+45eSTUZa
EenaOZUXQX9RVA2MEA4TuaIIAM7uMEbU2ta0zM8A9WathRkqxqNN/2l24Y3AjWek
xA1thE7wHvGtvhAO3v1S1GFbH/bbBoLSm3Ry+dZV8Hw+CK+h/soXzEjg7uIR67gW
SRZ3CPOGK2/0pQTLDMxQ9zCAzgAArMFAtjEe0Os51NgK5r170s00EY4mNTSE7285
5dLg+vJ3dcT5R1rbvElE3HI0JpmACNCGTxumML5f2GMiRgPyLsAbrOxDIhYessrj
QkZixmITW5dDvltbB/Rc8yojR3qvSe5SRD0kH/R2wikJnFA/rlQHWKR37e0/uMOu
cQGgeQB5EVF9Kskljo9VyPk7laqCJMMoZc1Ka21QhSLRDbNuXrNcfaDGDMJ5uk+w
3rDktpcFb/4cv5Jc+noMym+MiEZemvQz9cJjlBGdov/tPvjzaJERtbjrzSXQpO5f
C6CO4UwI/B/OEswbmNxW50Lh1rGQUrrVVSxpT2Co18xaAJO144cqkMO+UcDxWcgr
PFX6vkNXMsPZ4hxALqlxYZUCAwEAAQ==
-----END PUBLIC KEY-----`

var testPrivateKey string = `-----BEGIN PRIVATE KEY-----
MIIJRQIBADANBgkqhkiG9w0BAQEFAASCCS8wggkrAgEAAoICAQDJsc1WYHT7203J
nIqIwIlDPSBRb9C011B39dfbrDpIJC0lWgz6jqvm9vP1hCgodh7y995R0SYTv1Si
swyad0FLy/yUzBa4ITL0x9wxGxJafPy73JHKFV+NqoScwke7Y/DTkYuFBV/t8P7j
l5JNRloR6do5lRdBf1FUDYwQDhO5oggAzu4wRtTa1rTMzwD1Zq2FGSrGo03/aXbh
jcCNZ6TEDW2ETvAe8a2+EA7e/VLUYVsf9tsGgtKbdHL51lXwfD4Ir6H+yhfMSODu
4hHruBZJFncI84Yrb/SlBMsMzFD3MIDOAACswUC2MR7Q6znU2ArmvXvSzTQRjiY1
NITvbznl0uD68nd1xPlHWtu8SUTccjQmmYAI0IZPG6Ywvl/YYyJGA/IuwBus7EMi
Fh6yyuNCRmLGYhNbl0O+W1sH9FzzKiNHeq9J7lJEPSQf9HbCKQmcUD+uVAdYpHft
7T+4w65xAaB5AHkRUX0qySWOj1XI+TuVqoIkwyhlzUprbVCFItENs25es1x9oMYM
wnm6T7DesOS2lwVv/hy/klz6egzKb4yIRl6a9DP1wmOUEZ2i/+0++PNokRG1uOvN
JdCk7l8LoI7hTAj8H84SzBuY3FbnQuHWsZBSutVVLGlPYKjXzFoAk7XjhyqQw75R
wPFZyCs8Vfq+Q1cyw9niHEAuqXFhlQIDAQABAoICAQC9xm5ON7Paxh4K9R/kTETa
30jpVywo++7a8JaKOyMbfe58lp5fop5cU0B4YkDm0T2Nn2uvz/rj2cLo00+oh00I
5IZj+yPlXFd1uheUnMRIIBItMPx8CGBAC5F7bdHQn9iZOPjt0IDSgU9TFeqyit90
u3R5ea7IEeOUEqsW8CffInYlTI8RHZRp1FuJ2bwtKs9ZzLRS8pURHUqeL6Jdaoe9
cGT7eMq2UvAHRVS4u+KTsobrLHopRi6j1o3YRbPW8w/rXFYwRjbeIDSEkHIMOMm/
O0QFSB2WAWFPY5MqF5SXASwwqA/6fFtHEjDMoodnnV+ke+VmE25KllWc+i2anCz8
Rgy+MN+AfmQM+iAYwweybfOMh6DeaJvGiUOxy7ltgwQz6Kd3cbi03G4L9FsnQUnE
GWp36DZ7ivDOSVW1cFtxDpT5OFJPDkTqa7fG2iOOCoRsKPGefM3FN1UO0OnBr1sb
yDa0KwqLID/hBEPsiALcvJSSRhGahViIB1o3mYuW1U/oD0QaMaeHn8EvDP1NSdRh
+/HeUHejvbzuknr1ItNQ2JUXtvQjQfkUFeL8Pv7nxqLnZMZioTYemDYEwvH97JJy
/DL6yLYBnACGl3MPmlKZ0+8WFCMUpBMIhQGVTytj8ciFxLU+juzpF+O7km62wLXp
f/1q0ZBPSDOBBdIKXo8SgQKCAQEA6UyMLbZUrLptJIT2zKFQcaS2R+hvj3U40P0x
dVYiwZ6HdeiIkQFei4VMG1RqAAoTx/RUiFLMP/Ik9B+dSk3OYWaxyvyc0/5vYQvs
c9XxoGkjL0tbH4hNAkaGO7R8Cch3ZXRD3DOjVjgC2b5wQXXaNwq10yLUqG16tZPC
dg1s1xiY1v6y3I7opWncaLP27WqYSPz9WLrT+II+biFHeizGpbNpTyT7cqLMYNKc
+NGvnhYf5ry0PeV+8SwFC6OJN656Ped2fQPNKBCrK0ZbLVNqmWNZ60r8AGQ9zyBr
QI5Pa8jiH8whkf313nXJozo/AccwR3CDVJYObx1Ge+UGWTytdwKCAQEA3VH9l1Ri
/x5B15A3dVSLpYivtSxdTMgXEVwZIHjuo00MWdjkpruOuM7wRhYGx/zlR+ch7ki2
7Jb5ejaBnzpyU4nv1ekdF5ZrTv/yEP3oY1o8GHYWbuYgvEFSR7Gvq0xa+X0grPpp
J3aSEHYxhItrxZK48FTyv8hH0ltlqYg9wWWR01ocRZeKUt770QUF6V//qZU/i0Cx
2mzA2FY8xer1qDZyNkd2oTNHV8XDixEjDU3Ugs/7PBZLcfnZZt5TQB7Vm+MalfVU
40WydJ7gXjt3MCp8Os7NESkI6roaIBnKBC9gjoDtHcdGQDtMgfk5qiybT1dv4I9R
7eZqxANo8Rz8UwKCAQEA36qKrfyjG2Iz5xIuxqpVRE6kjzYRVpkMmphThWnKMpR4
zBreax7D9MEb4QvCAD2pD0d4j6XJufi9Yuq4UpdbqFfbVn9vH3NMdt8Gl1tipuaF
W/9D4mw7YFYatTzoujxd839O29sJ2kwit3zzhF6nkaOMBFrdRIiJX3HEuodOdL1Z
Mq3G7tt3wbZHIH6A2scaLseVYC7lr9e2YME5FLG+1Pe3m7AZ/aKEjML+yTHGy6ns
dbsuljTiyfbo82qa0C5Pde/l0h8F3kZ0xC1UlpTlmx78Ay/Ff96av2wWRzLog654
1AFRofS3dsq4QOxDocHE0IjW8A5Y0kBf5cIBnyYkKwKCAQEArnNUWpZPuhxVdd7d
eAR+VqqZJUuk90K4vBxGSgxIvjubQq7t/GbWpuGnDveJvNWgvL55RmPWnEKcvzJ7
ldDyENsnSwuGvPL5/rlMSwx4wud7eySJpjyLDRjIDG8IsxNqmkGBIhf4Dv0tQQCJ
I5rqBkASuo2bEoSB6FPWnH0hgHHZMilTI5/BjnTpNOaqtDvRQBC+l7sU9cDHeT7w
hGkh3cec2yAVaBcNYyglbFbDtFbm7X2W4NQJ//sa3DTelio34bpvWEia8tIbSkV5
QY3J8xNp/MjJZ39a4fpzYV4ymH3ntCv3u4M54qNbORAD3hlvCmk3bGBMCiOXgI3X
iEZ6tQKCAQEA14qOF3D+Zy3spZ60xKFlCZa1bFPvk9ab1Z50bJKQzinjyGzQbvDL
ABwDJMP6F1mQyjEADUODda9gYzHjpel5NaXx/8vj1i5Ydv1/JXQAcwxOQXDouDLw
dCJ3rxxgcTYlIGGbXIvJ9FCFMkDGxYsgsUbRtIVgKLCoOO9wQG9/AvEqI2LOE3OB
oyULFzV2S2IWZkUrSAjlykMtFa26nwi3xampM47sIL2igdnRTwh0MHgbWjvoh+Oj
sDaQRUrV98jKrkYfTMl5XhabzEWjXxVCPkyJTI0ylnx1jTsYF2R2gIebvVVsWeAN
2XPCDnhB9l68DFEkEcw2vpSZ3y5rPavYtg==
-----END PRIVATE KEY-----
`

func TestSealing(t *testing.T) {
	yamlData := `---
mail:
  accounts:
  - name: test1
    smtp:
      sender_address: "example@example.com"
      server_address: "smtp.example.com"	
      port: 465
      username: joeuser@example.com
      password: "it's a secret"
    imap:
      server_address: "imap.example.com"
      port: 993
      username: janeuser@example.com
      password: "it's a secret too"
  send_limits: []`

	yamldata := []byte(yamlData)
	config, err := parseAndValidateConfig(yamldata)
	assert.Nil(t, err)

	//check unsealed states
	assert.False(t, config.Mail.Accounts[0].SMTP.Password.IsValueSealed())
	assert.False(t, config.Mail.Accounts[0].IMAP.Password.IsValueSealed())

	//make a secret sealer
	sealer := secrets.MakeSecretSealer()
	err = sealer.LoadPublicKey([]byte(testPublicKey))
	assert.Nil(t, err)

	//seal the config
	err = config.Seal(sealer)
	assert.Nil(t, err)

	//check sealed states
	assert.True(t, config.Mail.Accounts[0].SMTP.Password.IsValueSealed())
	assert.True(t, config.Mail.Accounts[0].IMAP.Password.IsValueSealed())

	//validate the config again
	vp := validation.ValidationProcess{}
	err = config.Validate(&vp)
	assert.Nil(t, err)

	err = vp.GetFinalValidationError()
	assert.Nil(t, err)

	//unseal the sealed values and verify them
	unsealer := secrets.MakeSecretUnsealer()
	err = unsealer.LoadPrivateKey([]byte(testPrivateKey), "")
	assert.Nil(t, err)

	err = config.Unseal(unsealer)
	assert.Nil(t, err)

	assert.Equal(t, "it's a secret", config.Mail.Accounts[0].SMTP.Password.GetValue())
	assert.False(t, config.Mail.Accounts[0].SMTP.Password.IsValueSealed())

	assert.Equal(t, "it's a secret too", config.Mail.Accounts[0].IMAP.Password.GetValue())
	assert.False(t, config.Mail.Accounts[0].IMAP.Password.IsValueSealed())

}

type MockSealer struct {
}

func (ms *MockSealer) LoadPublicKeyFromFile(filename string) error {
	return nil
}
func (ms *MockSealer) LoadPublicKey(rawBytes []byte) error {
	return nil
}
func (ms *MockSealer) ClearKeys() {
	//do nothing
}
func (ms *MockSealer) GetMaximumSecretSize() (int, error) {
	return 10, nil
}
func (ms *MockSealer) SealSecret(secretToSeal string) (string, error) {
	return "", fmt.Errorf("mock sealer failed")
}
func (ms *MockSealer) SealSecretHolder(holder secrets.SecretHolder) {
	holder.Seal(ms)
}

func TestSealUnsealFailure(t *testing.T) {
	yamlData := `---
mail:
  accounts:
  - name: test1
    smtp:
      sender_address: "example@example.com"
      server_address: "smtp.example.com"	
      port: 465
      username: joeuser@example.com
      password: "it's a secret"
    imap:
      server_address: "imap.example.com"
      port: 993
      username: janeuser@example.com
      password: "sealed(+invalidseal==)"
  send_limits: []`

	yamldata := []byte(yamlData)
	config, err := parseAndValidateConfig(yamldata)
	assert.Nil(t, err)

	//check password states
	assert.False(t, config.Mail.Accounts[0].SMTP.Password.IsValueSealed())
	assert.Equal(t, "it's a secret", config.Mail.Accounts[0].SMTP.Password.GetValue())
	assert.True(t, config.Mail.Accounts[0].IMAP.Password.IsValueSealed())
	assert.Equal(t, "sealed(+invalidseal==)", config.Mail.Accounts[0].IMAP.Password.GetValue())

	//make a mock secret sealer so that the seal operation fails
	sealer := &MockSealer{}

	//seal the config
	err = config.Seal(sealer)
	assert.ErrorContains(t, err, "at path mail.accounts[0].SMTP.password")
	assert.ErrorContains(t, err, "failed to seal secret: mock sealer failed")

	//check password states
	assert.False(t, config.Mail.Accounts[0].SMTP.Password.IsValueSealed())
	assert.Equal(t, "it's a secret", config.Mail.Accounts[0].SMTP.Password.GetValue())
	assert.True(t, config.Mail.Accounts[0].IMAP.Password.IsValueSealed())
	assert.Equal(t, "sealed(+invalidseal==)", config.Mail.Accounts[0].IMAP.Password.GetValue())

	//because the test config has invalid sealed data, we can use the real unsealer in the test
	//it will fail on the invalid data
	unsealer := secrets.MakeSecretUnsealer()
	unsealer.LoadPrivateKey([]byte(testPrivateKey), "")

	//unseal the config
	err = config.Unseal(unsealer)
	assert.ErrorContains(t, err, "at path mail.accounts[0].IMAP.password")
	assert.ErrorContains(t, err, "failed to unseal secret crypto/rsa: decryption error")

	//check password states
	assert.False(t, config.Mail.Accounts[0].SMTP.Password.IsValueSealed())
	assert.Equal(t, "it's a secret", config.Mail.Accounts[0].SMTP.Password.GetValue())
	assert.True(t, config.Mail.Accounts[0].IMAP.Password.IsValueSealed())
	assert.Equal(t, "sealed(+invalidseal==)", config.Mail.Accounts[0].IMAP.Password.GetValue())

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
