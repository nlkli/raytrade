package app

import rl "github.com/gen2brain/raylib-go/raylib"

// Rect — базовый прямоугольник компонента (позиция + размер)
type Rect struct {
	p rl.Vector2 // верхний левый угол
	s rl.Vector2 // ширина и высота
}

// NewRect создаёт новый Rect
func NewRect(pX, pY, sW, sH float32) *Rect {
	return &Rect{
		p: rl.NewVector2(pX, pY),
		s: rl.NewVector2(sW, sH),
	}
}

// Fill рисует прямоугольник заливкой
func (r *Rect) Fill(c rl.Color) {
	rl.DrawRectangleV(r.p, r.s, c)
}

// Outline рисует обводку прямоугольника
func (r *Rect) Outline(thickness float32, c rl.Color) {
	rl.DrawRectangleLinesEx(rl.NewRectangle(r.p.X, r.p.Y, r.s.X, r.s.Y), thickness, c)
}

// Contains проверяет, находится ли точка (x, y) внутри Rect
func (r *Rect) Contains(x, y float32) bool {
	return x >= r.p.X && x <= r.p.X+r.s.X &&
		y >= r.p.Y && y <= r.p.Y+r.s.Y
}

// ContainsV проверяет, находится ли вектор v внутри Rect
func (r *Rect) ContainsV(v rl.Vector2) bool {
	return r.Contains(v.X, v.Y)
}

// Intersects проверяет пересечение с другим Rect
func (r *Rect) Intersects(other *Rect) bool {
	return r.p.X < other.p.X+other.s.X &&
		r.p.X+r.s.X > other.p.X &&
		r.p.Y < other.p.Y+other.s.Y &&
		r.p.Y+r.s.Y > other.p.Y
}

// Center возвращает центр Rect
func (r *Rect) Center() rl.Vector2 {
	return rl.NewVector2(r.p.X+r.s.X/2, r.p.Y+r.s.Y/2)
}

// BottomRight возвращает нижний правый угол
func (r *Rect) BottomRight() rl.Vector2 {
	return rl.NewVector2(r.p.X+r.s.X, r.p.Y+r.s.Y)
}

// MaxX возвращает правую границу Rect
func (r *Rect) MaxX() float32 { return r.p.X + r.s.X }

// MaxY возвращает нижнюю границу Rect
func (r *Rect) MaxY() float32 { return r.p.Y + r.s.Y }

// Move смещает Rect на dx, dy
func (r *Rect) Move(dx, dy float32) {
	r.p.X += dx
	r.p.Y += dy
}

// MoveTo перемещает Rect в точку (x, y)
func (r *Rect) MoveTo(x, y float32) {
	r.p.X = x
	r.p.Y = y
}

// Resize изменяет ширину и высоту на dw, dh
func (r *Rect) Resize(dw, dh float32) {
	r.s.X += dw
	r.s.Y += dh
}

// SetSize устанавливает ширину и высоту
func (r *Rect) SetSize(w, h float32) {
	r.s.X = w
	r.s.Y = h
}

// Shrink уменьшает Rect на px, py со всех сторон
func (r *Rect) Shrink(px, py float32) {
	r.p.X += px
	r.p.Y += py
	r.s.X -= 2 * px
	r.s.Y -= 2 * py
}

// Expand увеличивает Rect на px, py со всех сторон
func (r *Rect) Expand(px, py float32) {
	r.p.X -= px
	r.p.Y -= py
	r.s.X += 2 * px
	r.s.Y += 2 * py
}

// Clone создаёт копию Rect
func (r *Rect) Clone() *Rect {
	return &Rect{
		p: rl.NewVector2(r.p.X, r.p.Y),
		s: rl.NewVector2(r.s.X, r.s.Y),
	}
}

// Scale масштабирует размер Rect (позиция не меняется)
func (r *Rect) Scale(sx, sy float32) {
	r.s.X *= sx
	r.s.Y *= sy
}

// Lerp создаёт новый Rect, интерполированный между r и to по t [0..1]
func (r *Rect) Lerp(to *Rect, t float32) *Rect {
	return &Rect{
		p: rl.NewVector2(rl.Lerp(r.p.X, to.p.X, t), rl.Lerp(r.p.Y, to.p.Y, t)),
		s: rl.NewVector2(rl.Lerp(r.s.X, to.s.X, t), rl.Lerp(r.s.Y, to.s.Y, t)),
	}
}
