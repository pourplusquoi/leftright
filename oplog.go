package leftright

type operation int8

const (
	_ operation = iota
	operationInsert
	operationRemove
)

type oplog struct {
	op    operation
	key   interface{}
	value interface{}
}

func newInsert(key, value interface{}) oplog {
	return oplog{
		op:    operationInsert,
		key:   key,
		value: value,
	}
}

func newRemove(key interface{}) oplog {
	return oplog{
		op:  operationRemove,
		key: key,
	}
}
