package svg

// Theme defines the color palette for an SVG card.
type Theme struct {
	Name          string
	Background    string
	Surface       string
	Border        string
	PrimaryText   string
	SecondaryText string
	Accent        string
	MutedText     string
	ProgressBg    string
	IconColor     string
}

// Themes maps theme names to their color palettes.
var Themes = map[string]Theme{
	"dark": {
		Name:          "dark",
		Background:    "#111214",
		Surface:       "#1c2128",
		Border:        "#30363d",
		PrimaryText:   "#ffffff",
		SecondaryText: "#9ba1a8",
		Accent:        "#58a6ff",
		MutedText:     "#6e7681",
		ProgressBg:    "#21262d",
		IconColor:     "#9ba1a8",
	},
	"light": {
		Name:          "light",
		Background:    "#ffffff",
		Surface:       "#f6f8fa",
		Border:        "#d0d7de",
		PrimaryText:   "#1f2328",
		SecondaryText: "#59636e",
		Accent:        "#0969da",
		MutedText:     "#818b98",
		ProgressBg:    "#e1e4e8",
		IconColor:     "#59636e",
	},
	"github": {
		Name:          "github",
		Background:    "#ffffff",
		Surface:       "#f6f8fa",
		Border:        "#d0d7de",
		PrimaryText:   "#24292f",
		SecondaryText: "#57606a",
		Accent:        "#0969da",
		MutedText:     "#8c959f",
		ProgressBg:    "#e1e4e8",
		IconColor:     "#57606a",
	},
	"midnight": {
		Name:          "midnight",
		Background:    "#0a0e27",
		Surface:       "#141b3d",
		Border:        "#2d3561",
		PrimaryText:   "#c8d3f5",
		SecondaryText: "#7c8db5",
		Accent:        "#7c3aed",
		MutedText:     "#4a5578",
		ProgressBg:    "#1a2148",
		IconColor:     "#7c8db5",
	},
	"neon": {
		Name:          "neon",
		Background:    "#0d0221",
		Surface:       "#1a0933",
		Border:        "#3d1466",
		PrimaryText:   "#f0f0ff",
		SecondaryText: "#b388ff",
		Accent:        "#ff00ff",
		MutedText:     "#7b2cbf",
		ProgressBg:    "#26004d",
		IconColor:     "#b388ff",
	},
	"ocean": {
		Name:          "ocean",
		Background:    "#011627",
		Surface:       "#0d2d4a",
		Border:        "#1d3b53",
		PrimaryText:   "#d6deeb",
		SecondaryText: "#8badc4",
		Accent:        "#82aaff",
		MutedText:     "#5f7e97",
		ProgressBg:    "#0a2540",
		IconColor:     "#8badc4",
	},
	"sunset": {
		Name:          "sunset",
		Background:    "#1a0f0a",
		Surface:       "#2d1b12",
		Border:        "#4a2c1e",
		PrimaryText:   "#ffe8d6",
		SecondaryText: "#c9a88f",
		Accent:        "#ff7847",
		MutedText:     "#8a6b52",
		ProgressBg:    "#2a1a12",
		IconColor:     "#c9a88f",
	},
	"forest": {
		Name:          "forest",
		Background:    "#0f1a0f",
		Surface:       "#1a2b1a",
		Border:        "#2d4a2d",
		PrimaryText:   "#d4e8d4",
		SecondaryText: "#8cb88c",
		Accent:        "#56c568",
		MutedText:     "#5a7a5a",
		ProgressBg:    "#1c2c1c",
		IconColor:     "#8cb88c",
	},
	"dracula": {
		Name:          "dracula",
		Background:    "#282a36",
		Surface:       "#383a4a",
		Border:        "#44475a",
		PrimaryText:   "#f8f8f2",
		SecondaryText: "#9aa0c0",
		Accent:        "#bd93f9",
		MutedText:     "#6272a4",
		ProgressBg:    "#383a4a",
		IconColor:     "#9aa0c0",
	},
	"nord": {
		Name:          "nord",
		Background:    "#2e3440",
		Surface:       "#3b4252",
		Border:        "#434c5e",
		PrimaryText:   "#eceff4",
		SecondaryText: "#8896ab",
		Accent:        "#81a1c1",
		MutedText:     "#4c566a",
		ProgressBg:    "#3b4252",
		IconColor:     "#8896ab",
	},
}

// DefaultTheme is used when no theme is specified.
const DefaultTheme = "dark"

// GetTheme returns the theme for the given name, or an error if unknown.
func GetTheme(name string) (Theme, bool) {
	t, ok := Themes[name]
	return t, ok
}
