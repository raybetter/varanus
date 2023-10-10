package app

import (
	"fmt"
	"io"
	"strings"
)

type SealConfigArgs struct {
	Input          *string
	Output         *string
	ForceOverwrite *bool
	PublicKey      *string
}

func (c SealConfigArgs) HumanReadable() string {

	var sb strings.Builder

	fmt.Fprintln(&sb, "Sealing config")
	fmt.Fprintln(&sb, "  Input: ", *c.Input)
	fmt.Fprintln(&sb, "  PublicKey: ", *c.PublicKey)
	fmt.Fprintln(&sb, "  Output: ", *c.Output)
	fmt.Fprintln(&sb, "  ForceOverwrite: ", *c.ForceOverwrite)

	return sb.String()
}

type UnsealConfigArgs struct {
	Input          *string
	Output         *string
	ForceOverwrite *bool
	PrivateKey     *string
	Passphrase     *string
}

func (c UnsealConfigArgs) HumanReadable() string {

	var sb strings.Builder

	fmt.Fprintln(&sb, "Unsealing config")
	fmt.Fprintln(&sb, "  Input: ", *c.Input)
	fmt.Fprintln(&sb, "  PrivateKey: ", *c.PrivateKey)
	fmt.Fprintf(&sb, "  Passphrase: <redacted value of length %d>\n", len(*c.Passphrase))
	fmt.Fprintln(&sb, "  Output: ", *c.Output)
	fmt.Fprintln(&sb, "  ForceOverwrite: ", *c.ForceOverwrite)

	return sb.String()
}

type CheckConfigArgs struct {
	Input      *string
	PrivateKey *string
	Passphrase *string
}

func (c CheckConfigArgs) HumanReadable() string {

	var sb strings.Builder

	fmt.Fprintln(&sb, "Checking config with:")
	fmt.Fprintln(&sb, "  Input: ", *c.Input)
	fmt.Fprintln(&sb, "  PrivateKey: ", *c.PrivateKey)
	fmt.Fprintf(&sb, "  Passphrase: <redacted value of length %d>\n", len(*c.Passphrase))

	return sb.String()
}

type VaranusApp interface {
	SealConfig(args *SealConfigArgs, outputStream io.Writer) error
	UnsealConfig(args *UnsealConfigArgs, outputStream io.Writer) error
	CheckConfig(args *CheckConfigArgs, outputStream io.Writer) error
}

type ApplicationError struct {
	theError error
}

func (ae ApplicationError) Error() string {
	return ae.theError.Error()
}

func (ae *ApplicationError) Unwrap() error { return ae.theError }
