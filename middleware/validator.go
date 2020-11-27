package middleware

type IValidator interface {
	Validate() error
}
