package app

type config struct {
	Theme theme `json:"theme"`
}

type theme struct {
	Name    string      `json:"name"`
	IsLight bool        `json:"is_light"`
	Colors  colorScheme `json:"colors"`
}

type colorScheme struct {
	Background    [5]string       `json:"background"`
	Foreground    [4]string       `json:"foreground"`
	Selection     selectionColors `json:"selection"`
	Cursor        cursorColors    `json:"cursor"`
	Base          ansiColors      `json:"base"`
	Bright        ansiColors      `json:"bright"`
	Dim           ansiColors      `json:"dim"`
	Diff          diffColors      `json:"diff"`
	CodeSelection [2]string       `json:"code_selection"`
	Comment       string          `json:"comment"`
}

type selectionColors struct {
	Bg string `json:"bg"`
	Fg string `json:"fg"`
}

type cursorColors struct {
	Bg string `json:"bg"`
	Fg string `json:"fg"`
}

type ansiColors struct {
	Black   string `json:"black"`
	Red     string `json:"red"`
	Green   string `json:"green"`
	Yellow  string `json:"yellow"`
	Blue    string `json:"blue"`
	Magenta string `json:"magenta"`
	Cyan    string `json:"cyan"`
	White   string `json:"white"`
	Orange  string `json:"orange"`
	Pink    string `json:"pink"`
}

type diffColors struct {
	Add    string `json:"add"`
	Delete string `json:"delete"`
	Change string `json:"change"`
	Text   string `json:"text"`
}
