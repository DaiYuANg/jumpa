// backend 模拟真实后端：configx + logx + eventx + httpx + dix + dbx(sqlite/mysql/postgres)
//
// 运行: go run ./backend
// 环境变量: APP_SERVER_PORT=3000, APP_DB_DRIVER=sqlite, APP_DB_DSN=file:app.db
package main

import "github.com/DaiYuANg/arcgo-rbac-template/app"

func main() {
	app.Run()
}
