package edb

type Image struct {
	DBModel

	URI string
}

func (i Image) Namespace() string {
	return "image"
}
