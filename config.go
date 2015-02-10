package slack

type Config struct {
	Ok bool
	Error string
	Url string
	Args Foo
}

type Foo struct {
	Token string
}