package die_test

import (
	"fmt"
	"testing"

	die "."
)

func TestRaiseMessage(t *testing.T) {
	tests := []struct {
		val int
		exp string
	}{
		{0, "0"},
		{3, "three!"},
		{5, "5"},
	}

	for _, test := range tests {
		ret := convertIfThree(test.val)
		if ret != test.exp {
			t.Errorf("want %v, get %v", test.exp, ret)
		}
	}
}

func convertIfThree(n int) (ret string) {
	defer die.Revival(func(msg interface{}) {
		str := msg.(string)
		ret = str
	})

	die.If(n == 3, fmt.Sprintf("three!"))
	return fmt.Sprintf("%d", n)
}

func TestRaiseError(t *testing.T) {
	tests := []struct {
		val int
		exp error
	}{
		{0, nil},
		{3, fmt.Errorf("three!")},
		{5, nil},
	}

	for _, test := range tests {
		ret := failIfThree(test.val)
		if !errorEQ(ret, test.exp) {
			t.Errorf("want %v, get %v", test.exp, ret)
		}
	}
}

func failIfThree(n int) (err error) {
	defer die.RevivalErr(func(e error) {
		err = e
	})

	if n == 3 {
		err = fmt.Errorf("three!")
	}
	die.IfErr(err)
	return nil
}

func errorEQ(e1, e2 error) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 != nil && e2 != nil && e1.Error() == e2.Error() {
		return true
	}
	return false
}

func TestRaiseMessageDetail(t *testing.T) {
	tests := []struct {
		val int
		exp string
	}{
		{0, "0"},
		{3, "three! line:102"},
		{5, "5"},
	}

	for _, test := range tests {
		ret := convertIfThreeDetail(test.val)
		if ret != test.exp {
			t.Errorf("want %v, get %v", test.exp, ret)
		}
	}
}

func convertIfThreeDetail(n int) (ret string) {
	defer die.RevivalDetail(func(_ string, line int, msg interface{}) {
		str := msg.(string)
		ret = fmt.Sprintf("%s line:%d", str, line)
	})

	die.If(n == 3, fmt.Sprintf("three!"))
	return fmt.Sprintf("%d", n)
}
