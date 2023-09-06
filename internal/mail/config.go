package mail

import (
	"fmt"

	"github.com/spf13/viper"
)

const KeyMailConfiguration = "mail_configuration"

type SMTPConfig struct {
	SenderAddress string
	ServerAddress string
	Port          int
	Username      string
	Password      string
}

type IMAPConfig struct {
	ServerAddress string
	Port          int
	Username      string
	Password      string
}

type ServerConfig struct {
	SMTP *SMTPConfig
	IMAP *IMAPConfig
}

type SendLimit struct {
	SendLimit int
	Accounts  []string
}

type MailConfig struct {
	Accounts   map[string]ServerConfig
	SendLimits []SendLimit `mapstructure:"send_limits"`
}

func PrintConfig(mailconfig MailConfig) {

	fmt.Println("Accounts")
	for accountName, accountConfig := range mailconfig.Accounts {
		fmt.Printf("  - %s\n", accountName)
		fmt.Printf("    SMTP config: %v\n", accountConfig.SMTP)
		fmt.Printf("    IMAP config: %v\n", accountConfig.IMAP)
		fmt.Println()
	}
	fmt.Println("Send Limits")
	for _, sendLimit := range mailconfig.SendLimits {
		fmt.Printf("  - %d for %v\n", sendLimit.SendLimit, sendLimit.Accounts)
	}
}

func LoadMailConfig(viperConfig *viper.Viper) (MailConfig, error) {

	mailConfig := viperConfig.Sub(KeyMailConfiguration)

	viper.SetDefault("port", 333)

	var mail_config MailConfig
	err := mailConfig.Unmarshal(&mail_config)

	if err != nil {
		return MailConfig{}, err
	}
	return mail_config, nil
}
