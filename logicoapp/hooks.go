package logicoapp

import "github.com/gnemade360/go-boot/logicoapp/internal/hooks"

func SimpleHook(f func()) Hook {
	return hooks.SimpleHook(f)
}
