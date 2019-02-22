package stash

//Command ...
type Command interface {
	/*To create and publish stash for the first time*/
	Create(name string, hash []byte) error
	Fetch(name string) error
	Publish(name string) error
	/*Stash maintainence*/
	Pull() error
	Push() error
	Encrypt() error
	Decrypt() (map[string][]byte, error)
	/*CRUD*/
	Get(pattern string) ([]string, error)
	AddOrUpdate(key string, value []byte) error
	Delete(key string) error
	GetHash() ([]byte, error)
}

//Store ...
type Store interface {
	//Read returns data as a string
	Read(name string) ([]byte, error)
	//Write writes json
	Write(string, []byte) error
}
