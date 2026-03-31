package cli

import "fmt"

const banner = `
    ╲╲╲╲╲╲╲╲╲
     ╲╲╲╲╲╲╲╲──────╮
      ╲╲╲╲╲╲╲──────┤
       ╲╲╲╲╲╲──────┤   ╭──────────────╮
        ╲╲╲╲╲──────┼───┤  rice-rail   │
       ╱╱╱╱╱╱──────┤   ╰──────────────╯
      ╱╱╱╱╱╱╱──────┤
     ╱╱╱╱╱╱╱╱──────╯
    ╱╱╱╱╱╱╱╱╱
`

const bannerCompact = `  ══╦══ rice-rail ══╦══  project convergence toolkit`

func printBanner() {
	fmt.Print(banner)
}
