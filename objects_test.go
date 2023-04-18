package poly

type TypeString struct {
	ValueA string
}

type TypeFloat struct {
	ValueB float32
}

type TypeInt struct {
	ValueC int
	index  int
}

func (t *TypeInt) SetIndex(i int) {
	t.index = i
}

func (t *TypeInt) GetIndex() int {
	return t.index
}

type SlicesABC struct {
	TypeString []TypeString
	TypeBravo  []TypeFloat `poly:"TypeFloat"`
	TypeInt    TypeInt
	TypeIntP   *TypeInt
}
