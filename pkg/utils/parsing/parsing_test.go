package parsing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetKeyPair(t *testing.T) {
	r := require.New(t)

	str := []byte("name=arpg")
	key, val := GetKeyPair(str)
	r.Equal("name", key)
	r.Equal("arpg", val)

	str = []byte(" name = arpg")
	key, val = GetKeyPair(str)
	r.Equal("name", key)
	r.Equal("arpg", val)

	str = []byte(" name =")
	key, val = GetKeyPair(str)
	r.Equal("", key)
	r.Equal("", val)
}

func Test_PopFirstString(t *testing.T) {
	r := require.New(t)

	str := "arpg|game"
	firstStr, str := PopFirstString(str, "")
	r.Equal("", str)
	r.Equal("arpg|game", firstStr)

	str = "arpg,game;maker"
	firstStr, str = PopFirstString(str, "")
	r.Equal("game;maker", str)
	r.Equal("arpg", firstStr)

	str = "arpg;game,maker"
	firstStr, str = PopFirstString(str, "")
	r.Equal("game,maker", str)
	r.Equal("arpg", firstStr)

	str = "arpg,game|maker"
	firstStr, str = PopFirstString(str, "|")
	r.Equal("maker", str)
	r.Equal("arpg,game", firstStr)

}

func Test_PopFirstInt(t *testing.T) {
	r := require.New(t)

	str := "7.37.day"
	firstNum, str := PopFirstInt(str, ".")
	r.Equal("37.day", str)
	r.Equal(7, firstNum)
}

func Test_TryParseValue(t *testing.T) {
	r := require.New(t)

	var outputBool interface{} = false
	r.Equal(true, TryParseValue("1", &outputBool))
	r.Equal(true, outputBool)

	var outputInt8 interface{} = int8(0)
	r.Equal(true, TryParseValue("37", &outputInt8))
	r.Equal((int8)(37), outputInt8)

	var outputInt16 interface{} = int16(0)
	r.Equal(true, TryParseValue("-37", &outputInt16))
	r.Equal((int16)(-37), outputInt16)

	var outputInt interface{} = 0
	r.Equal(true, TryParseValue("37", &outputInt))
	r.Equal(37, outputInt)

	var outputUint8 interface{} = uint8(0)
	r.Equal(true, TryParseValue("37", &outputUint8))
	r.Equal((uint8)(37), outputUint8)

	var outputUint16 interface{} = uint16(0)
	r.Equal(true, TryParseValue("37", &outputUint16))
	r.Equal((uint16)(37), outputUint16)

	var outputUint interface{} = uint(0)
	r.Equal(true, TryParseValue("37", &outputUint))
	r.Equal(uint(37), outputUint)

	var outputFloat32 interface{} = float32(0)
	r.Equal(true, TryParseValue("3.1415", &outputFloat32))
	r.Equal((float32)(3.1415), outputFloat32)

	var outputFloat64 interface{} = float64(0)
	r.Equal(true, TryParseValue("3.1415", &outputFloat64))
	r.Equal((float64)(3.1415), outputFloat64)

	var outputStr interface{} = ""
	r.Equal(true, TryParseValue("dd3.1415", &outputStr))
	r.Equal("dd3.1415", outputStr)

}

func Test_GetSectionTitle(t *testing.T) {
	r := require.New(t)

	str := "[name]"
	r.Equal("name", GetSectionTitle(str))

	str = " [name ] "
	r.Equal("name ", GetSectionTitle(str))

}

func Test_ToDuration(t *testing.T) {
	r := require.New(t)

	r.Equal(2220, ToDuration("37s", 60))
	r.Equal(2220, ToDuration("37 s", 60))
	r.Equal(4, ToDuration("60 ms", 60))
	r.Equal(1, ToDuration("17ms", 60))
}
