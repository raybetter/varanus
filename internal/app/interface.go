package app

type SealConfigArgs struct {
	Input          *string
	Output         *string
	ForceOverwrite *bool
	PublicKey      *string
}

type UnsealConfigArgs struct {
	Input          *string
	Output         *string
	ForceOverwrite *bool
	PrivateKey     *string
	Passphrase     *string
}

type CheckConfigArgs struct {
	Input      *string
	PrivateKey *string
	Passphrase *string
}

type VaranusApp interface {
	SealConfig(args *SealConfigArgs) error
	UnsealConfig(args *UnsealConfigArgs) error
	CheckConfig(args *CheckConfigArgs) error
}

type ApplicationError struct {
	theError error
}

func (ae ApplicationError) Error() string {
	return ae.theError.Error()
}

func (ae *ApplicationError) Unwrap() error { return ae.theError }
