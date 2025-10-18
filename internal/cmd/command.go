package cmd

// command holds information about the application commands.
type command struct {
	Name  string
	Usage string
}

// commandList is a map of commands and their information.
type commandList map[string]command

var commands = commandList{
	"ghost": {Name: "ghost", Usage: "send ghost a prompt"},
}
