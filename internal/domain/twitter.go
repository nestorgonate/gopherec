package domain

import "context"

type TwitterAPI interface {
	Post(c context.Context, text string) (string, error)
}
