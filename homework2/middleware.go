package homework2

type Middleware func(next HandleFunc) HandleFunc
