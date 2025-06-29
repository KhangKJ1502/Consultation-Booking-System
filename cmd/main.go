package main

import (
	"cbs_backend/global"
	"cbs_backend/internal/initialize"
)

func main() {
	r := initialize.Run()

	// 🔥 Quan trọng: Chạy server tại đây
	if err := r.Run(global.ConfigConection.ServerCF.Host + ":" + global.ConfigConection.ServerCF.Port); err != nil {
		panic(err)
	}
}
