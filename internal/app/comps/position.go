
package comps

import "nlkli/raytrade/internal/app/core"

type Position struct {
	*Rect
}

func (p *Position) R() *Rect {
	return p.Rect
}

func (p *Position) Render(s *core.State) {

}
