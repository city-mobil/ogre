package proxy

import "context"

type Proxy interface {
	Start() <-chan error
	Stop(context.Context) error
}
