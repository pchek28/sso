package models

type User struct {
	ID       int64
	Enail    string
	PassHash []byte
}
