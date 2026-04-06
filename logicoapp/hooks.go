package logicoapp

import "github.com/gnemade360/logicoapp/logicoapp/internal/hooks"

func SimpleHook(f func()) Hook {
	return hooks.SimpleHook(f)
}
