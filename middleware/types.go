package middleware

type TransformerRegistry map[string]func(interface{}) (interface{}, error)
