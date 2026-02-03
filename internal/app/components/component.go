package components

import "nlkli/raytrade/internal/app/state"

type Component interface {
	Render(*state.State) state.Action
}
