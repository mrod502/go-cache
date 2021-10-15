package gocache

type Item interface {
	//Create creates any resources that relate to an initialized Item
	Create() error
	//Destroy removes any resources created by Item.Create()
	Destroy() error
}

type DB interface {
	Get(string) (Object, error)
	Put(string, Object) error
	Exists(string) (bool, error)
	Delete(string) error
	Where(Matcher) ([]Object, error)
	Keys() []string
}

type Dispatcher func(Item) error
