// backend 模拟真实后端：configx + logx + eventx + gocron + httpx + dix + dbx(sqlite/mysql/postgres)
//
// 运行: go run ./backend
// 环境变量:
// APP_SERVER_PORT=3000
// APP_DB_DRIVER=sqlite
// APP_DB_DSN=file:app.db
// APP_SCHEDULER_ENABLED=true
// APP_SCHEDULER_HEARTBEAT_SEC=60
// APP_SCHEDULER_DISTRIBUTED_ENABLED=false
// APP_SCHEDULER_DISTRIBUTED_KEY_PREFIX=gocron:lock
// APP_SCHEDULER_DISTRIBUTED_TTL_SEC=30
// APP_VALKEY_ENABLED=false
// APP_VALKEY_ADDR=127.0.0.1:6379
// APP_VALKEY_PASSWORD=
// APP_VALKEY_DB=0
// APP_VALKEY_USE_TLS=false
package main

func main() {
	Run()
}
