package theme

import "github.com/egoist/mygo/ui"

// DarkStudioTheme returns the Tactile Matte Pro dark theme.
func DarkStudioTheme() *ui.Theme {
	th := ui.DarkTheme()
	th.Background = ui.Hex("#0D1117")    // Deep matte obsidian
	th.Surface = ui.Hex("#161B22")       // Elevated matte studio card surface
	th.SurfaceHover = ui.Hex("#21262D")  // Tactile hover state
	th.SurfacePressed = ui.Hex("#1C2128")
	th.Border = ui.Hex("#30363D")        // Crisp precision 1px border
	th.Text = ui.Hex("#F0F6FC")          // Clean high-contrast text
	th.TextMuted = ui.Hex("#8B949E")     // Slate muted secondary text
	th.Accent = ui.Hex("#2F81F7")        // Precision cobalt blue
	th.AccentHover = ui.Hex("#388BFD")   // Tactile responsive hover
	th.AccentPressed = ui.Hex("#1F6FEB") // Physical press state
	th.AccentText = ui.Hex("#FFFFFF")
	th.Success = ui.Hex("#2EA043")       // Emerald status indicator
	th.Danger = ui.Hex("#F85149")        // Rose alert indicator
	th.Warning = ui.Hex("#D29922")       // Amber warning
	th.Radius = 6                        // Subtle engineering radius
	return th
}

// LightStudioTheme returns the clean precision light studio theme.
func LightStudioTheme() *ui.Theme {
	th := ui.LightTheme()
	th.Background = ui.Hex("#F6F8FA")    // Subtle neutral background
	th.Surface = ui.Hex("#FFFFFF")       // Clean studio cards
	th.SurfaceHover = ui.Hex("#F3F4F6")
	th.SurfacePressed = ui.Hex("#E5E7EB")
	th.Border = ui.Hex("#D0D7DE")        // Crisp technical border
	th.Text = ui.Hex("#1F2328")          // Deep slate charcoal text
	th.TextMuted = ui.Hex("#656D76")     // Slate muted text
	th.Accent = ui.Hex("#0969DA")        // Clean precision blue
	th.AccentHover = ui.Hex("#0860CA")
	th.AccentPressed = ui.Hex("#054DA7")
	th.AccentText = ui.Hex("#FFFFFF")
	th.Success = ui.Hex("#1A7F37")
	th.Danger = ui.Hex("#CF222E")
	th.Warning = ui.Hex("#9A6700")
	th.Radius = 6
	return th
}
