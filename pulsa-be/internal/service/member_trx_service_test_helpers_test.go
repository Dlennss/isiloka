package service

func int64Ptr(v int64) *int64 {
	return &v
}

type testErr string

func (e testErr) Error() string { return string(e) }

func assertErr(msg string) error { return testErr(msg) }
