// Code generated by "enumer -type=MarshalingType -trimprefix=MarshalingType -transform=snake"; DO NOT EDIT.

package keysutil

import (
	"fmt"
)

const _MarshalingTypeName = "asn1jws"

var _MarshalingTypeIndex = [...]uint8{0, 4, 7}

func (i MarshalingType) String() string {
	i -= 1
	if i >= MarshalingType(len(_MarshalingTypeIndex)-1) {
		return fmt.Sprintf("MarshalingType(%d)", i+1)
	}
	return _MarshalingTypeName[_MarshalingTypeIndex[i]:_MarshalingTypeIndex[i+1]]
}

var _MarshalingTypeValues = []MarshalingType{1, 2}

var _MarshalingTypeNameToValueMap = map[string]MarshalingType{
	_MarshalingTypeName[0:4]: 1,
	_MarshalingTypeName[4:7]: 2,
}

// MarshalingTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func MarshalingTypeString(s string) (MarshalingType, error) {
	if val, ok := _MarshalingTypeNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to MarshalingType values", s)
}

// MarshalingTypeValues returns all values of the enum
func MarshalingTypeValues() []MarshalingType {
	return _MarshalingTypeValues
}

// IsAMarshalingType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i MarshalingType) IsAMarshalingType() bool {
	for _, v := range _MarshalingTypeValues {
		if i == v {
			return true
		}
	}
	return false
}
