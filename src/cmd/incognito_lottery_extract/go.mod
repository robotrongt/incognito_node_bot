module github.com/robotrongt/incognito_node_bot/src/cmd/incognito_lottery_extract

go 1.15

require (
	github.com/mattn/go-sqlite3 v1.14.4
	github.com/robotrongt/incognito_node_bot/src/models v0.0.0-00010101000000-000000000000
)

replace github.com/robotrongt/incognito_node_bot/src/models => ../../models
