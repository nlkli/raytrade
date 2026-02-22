package comps

import rl "github.com/gen2brain/raylib-go/raylib"

type Rect struct {
	p rl.Vector2
	s rl.Vector2
}

func NewRect(pX, pY, sW, sH float32) *Rect {
	return &Rect{
		p: rl.NewVector2(pX, pY),
		s: rl.NewVector2(sW, sH),
	}
}

func (r *Rect) Fill(c rl.Color) {
	rl.DrawRectangleV(r.p, r.s, c)
}

func (r *Rect) Outline(thickness float32, c rl.Color) {
	rl.DrawRectangleLinesEx(rl.NewRectangle(r.p.X, r.p.Y, r.s.X, r.s.Y), thickness, c)
}

func (r *Rect) Contains(x, y float32) bool {
	return x >= r.p.X && x <= r.p.X+r.s.X &&
		y >= r.p.Y && y <= r.p.Y+r.s.Y
}

func (r *Rect) ContainsV(v rl.Vector2) bool {
	return r.Contains(v.X, v.Y)
}

func (r *Rect) Intersects(other *Rect) bool {
	return r.p.X < other.p.X+other.s.X &&
		r.p.X+r.s.X > other.p.X &&
		r.p.Y < other.p.Y+other.s.Y &&
		r.p.Y+r.s.Y > other.p.Y
}

func (r *Rect) Center() rl.Vector2 {
	return rl.NewVector2(r.p.X+r.s.X/2, r.p.Y+r.s.Y/2)
}

func (r *Rect) BottomRight() rl.Vector2 {
	return rl.NewVector2(r.p.X+r.s.X, r.p.Y+r.s.Y)
}

func (r *Rect) MaxX() float32 { return r.p.X + r.s.X }

func (r *Rect) MaxY() float32 { return r.p.Y + r.s.Y }

func (r *Rect) Move(dx, dy float32) {
	r.p.X += dx
	r.p.Y += dy
}

func (r *Rect) MoveTo(x, y float32) {
	r.p.X = x
	r.p.Y = y
}

func (r *Rect) Resize(dw, dh float32) {
	r.s.X += dw
	r.s.Y += dh
}

func (r *Rect) SetSize(w, h float32) {
	r.s.X = w
	r.s.Y = h
}

func (r *Rect) Shrink(px, py float32) {
	r.p.X += px
	r.p.Y += py
	r.s.X -= 2 * px
	r.s.Y -= 2 * py
}

func (r *Rect) Expand(px, py float32) {
	r.p.X -= px
	r.p.Y -= py
	r.s.X += 2 * px
	r.s.Y += 2 * py
}

func (r *Rect) Clone() *Rect {
	return &Rect{
		p: rl.NewVector2(r.p.X, r.p.Y),
		s: rl.NewVector2(r.s.X, r.s.Y),
	}
}

func (r *Rect) Scale(sx, sy float32) {
	r.s.X *= sx
	r.s.Y *= sy
}

func (r *Rect) SplitV(w float32) (*Rect, *Rect) {
	return &Rect{
			p: rl.Vector2{X: r.p.X, Y: r.p.Y},
			s: rl.Vector2{X: w, Y: r.s.Y},
		},
		&Rect{
			p: rl.Vector2{X: r.p.X + w, Y: r.p.Y},
			s: rl.Vector2{X: r.s.X - w, Y: r.s.Y},
		}
}

func (r *Rect) SplitH(h float32) (*Rect, *Rect) {
	return &Rect{
			p: rl.Vector2{X: r.p.X, Y: r.p.Y},
			s: rl.Vector2{X: r.s.X, Y: h},
		},
		&Rect{
			p: rl.Vector2{X: r.p.X, Y: r.p.Y + h},
			s: rl.Vector2{X: r.s.X, Y: r.s.Y - h},
		}
}
