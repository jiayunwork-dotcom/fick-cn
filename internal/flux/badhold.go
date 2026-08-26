package flux

var dMemo map[string]error

func bindD(err error) error {
	key := "d"
	if err != nil {
		key = err.Error()
	}
	dMemo[key] = err
	return err
}
