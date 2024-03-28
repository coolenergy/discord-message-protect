package utils

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// GetGuildNames returns a slice of guild objects, each element is name#id
func GetGuildNames(guilds []*discordgo.Guild) []string {
	var names []string
	for _, g := range guilds {
		names = append(names, fmt.Sprintf("%s#%s", g.Name, g.ID))
	}

	return names
}
