// Code generated by "stringer -type=Collection"; DO NOT EDIT.

package api

import "strconv"

const _Collection_name = "useraccounthashworkerhashworkertokenkeysgroupcalllistworkertasklistpaymentspost"

var _Collection_index = [...]uint8{0, 4, 15, 25, 31, 36, 40, 45, 53, 67, 75, 79}

func (i Collection) String() string {
	if i < 0 || i >= Collection(len(_Collection_index)-1) {
		return "Collection(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Collection_name[_Collection_index[i]:_Collection_index[i+1]]
}
