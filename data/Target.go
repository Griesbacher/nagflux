package data

//Target connects a datatype to a name
type Target struct {
	Name     string
	Datatype Datatype
}

func (t Target) String() string {
	return t.Name + "-" + string(t.Datatype)
}
