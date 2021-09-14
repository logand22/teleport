package teleport

import "embed"

//go:embed embed.assets
var embedFS embed.FS

func EmbedFS() embed.FS {
	return embedFS
}
