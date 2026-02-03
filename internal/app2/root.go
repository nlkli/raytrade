package app2

const (
	RH  float32 = 16        // Row height
	CLH float32 = RH * 2.33 // Command line height
)

type Root struct {
	*Rect
	parent *Rect

	f *Footer
}

func InitRoot() *Root {
	root := &Root{
		Rect: &Rect{},
	}
	root.f = &Footer{
		Rect: &Rect{},
		parent: root.Rect,
	}
	root.f.sl = &StatusLine{
		Rect: &Rect{},
		parent: root.f.Rect,
	}
	root.f.cl = &CommandLine{
		Rect: &Rect{},
		parent: root.f.Rect,
	}

	return root 
}

func (r *Root) Render(s *State) {
	r.f.Render(s)
}
