//go:generate stringer -type=Collection

package api

type Collection int

const (
	user Collection = iota
	accounthash
	workerhash
	worker
	token
	keys
	group
	calllist
	workertasklist
	payments
	post
)
