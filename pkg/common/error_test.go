package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_error_equal(t *testing.T) {
	r := require.New(t)

	err1 := Err_normal_exit
	err2 := Err_no_fallbackmod

	r.Equal(true, err1 == Err_normal_exit)
	r.Equal(true, err2 == Err_no_fallbackmod)

	m := map[string]int{}
	m[""] = 37

	for k, v := range m {
		fmt.Println(k, v)
	}
}
