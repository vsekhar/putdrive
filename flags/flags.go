package flags

import "flag"

//var UserToken = flag.String("usertoken", "", "OAUTH2 user token")
var ShowTokens = flag.Bool("show-tokens", false, "show OAUTH tokens")
var ItemIds = flag.String("item-ids", "", "comma-separated list of put.io file/folder IDs to move")
var Copy = flag.Bool("copy", true, "copy files")
var Delete = flag.Bool("delete", false, "delete files")

func init() {
	flag.Parse()
}
