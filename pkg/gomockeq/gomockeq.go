package gomockeq

import (
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

type eq struct {
	exp  interface{}
	opts []cmp.Option
}

// Eq returns new match to gomock with go-cmp integration.
func Eq(exp interface{}, opts ...cmp.Option) gomock.Matcher {
	return &eq{exp: exp, opts: opts}
}

func (e *eq) Matches(act interface{}) bool {
	return cmp.Equal(e.exp, act, e.opts...)
}

func (e *eq) String() string {
	return fmt.Sprintf("is of type %v", e.exp)
}
