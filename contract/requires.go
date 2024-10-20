package contract

type CheckFn func() error

func Require(validators ...CheckFn) error {
	for _, v := range validators {
		if err := v(); err != nil {
			return err
		}
	}
	return nil
}

func Must(validators ...CheckFn) {
	err := Require(validators...)
	if err != nil {
		panic(err)
	}
}
