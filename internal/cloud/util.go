package cloud

import "unicode"

var GSM7CharacterSet = map[rune]byte{
	'@': 0x00,
	'£': 0x01,
	'$': 0x02,
	'¥': 0x03,
	'_': 0x11,
	'!': 0x21,
	'#': 0x23,
	'%': 0x25,
	'&': 0x26,
	'(': 0x28,
	')': 0x29,
	'*': 0x2A,
	'+': 0x2B,
	',': 0x2C,
	'-': 0x2D,
	'.': 0x2E,
	'/': 0x2F,
	':': 0x3A,
	';': 0x3B,
	'<': 0x3C,
	'=': 0x3D,
	'>': 0x3E,
	'?': 0x3F,
}

func ToTitle(s string) string {
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}

func ToGSM7Bytes(b []byte) []byte {
	var result []byte
	for _, c := range b {
		if v, ok := GSM7CharacterSet[rune(c)]; ok {
			result = append(result, v)
		} else {
			result = append(result, c)
		}
	}
	return result
}
