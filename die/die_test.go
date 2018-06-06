package die

import (
	"fmt"
	"path/filepath"
	"testing"
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
	defer Revival(func(msg interface{}) {
		str := msg.(string)
		ret = str
	})

	If(n == 3, fmt.Sprintf("three!"))
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
	defer RevivalErr(func(e error) {
		err = e
	})

	if n == 3 {
		err = fmt.Errorf("three!")
	}
	IfErr(err)
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
		{3, "three! line:101"},
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
	defer RevivalDetail(func(_ string, line int, msg interface{}) {
		str := msg.(string)
		ret = fmt.Sprintf("%s line:%d", str, line)
	})

	If(n == 3, fmt.Sprintf("three!"))
	return fmt.Sprintf("%d", n)
}

func ExampleIf() {
	defer Revival(func(msg interface{}) {
		fmt.Println(msg)
	})

	If(0 == 1, fmt.Sprintf("0 == 1 == true"))
	If(0 != 1, fmt.Sprintf("0 != 1 == true"))
	// Output:
	// 0 != 1 == true
}

func ExampleIfErr() {
	defer RevivalErr(func(err error) {
		fmt.Println(err)
	})

	IfErr(nil)
	IfErr(fmt.Errorf("error message!"))
	// Output:
	// error message!
}

func ExampleRevival() {
	defer Revival(func(msg interface{}) {
		fmt.Println(msg)
	})

	If(0 != 1, fmt.Sprintf("0 != 1 == true"))
	// Output:
	// 0 != 1 == true
}

func ExampleRevivalErr() {
	defer RevivalErr(func(err error) {
		fmt.Println(err)
	})

	IfErr(fmt.Errorf("error message!"))
	// Output:
	// error message!
}

func ExampleRevivalDetail() {
	defer RevivalDetail(func(file string, line int, msg interface{}) {
		fmt.Println(filepath.Base(file))
		fmt.Println(line)
		fmt.Println(msg)
	})

	If(0 != 1, fmt.Sprintf("error message"))
	// Output:
	// die_test.go
	// 154
	// error message
}
