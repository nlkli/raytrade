package comps

import "nlkli/raytrade/internal/app/core"

type Order struct {
	*Rect
}

func (o *Order) R() *Rect {
	return o.Rect
}

func (o *Order) Render(s *core.State) {

}
