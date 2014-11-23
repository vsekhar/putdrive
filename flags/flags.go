package flags

import "flag"

var ItemIds = flag.String("item-ids", "", "comma-separated list of put.io file/folder IDs to move")
var Copy = flag.Bool("copy", true, "copy files")
var Delete = flag.Bool("delete", false, "delete files")

var NewTokens = flag.Bool("new-tokens", false, "fetch new OAUTH tokens")
var ShowTokens = flag.Bool("show-tokens", false, "show OAUTH tokens")


func init() {
	flag.Parse()
}
