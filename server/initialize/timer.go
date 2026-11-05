package initialize

import (
	"fmt"

	"optimus/server/global"
	"optimus/server/task"

	"github.com/robfig/cron/v3"
)

// Timer registers scheduled jobs, as GVA's initialize.Timer.
func Timer() {
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())
		_, err := global.OPTIMUS_Timer.AddTaskByFunc("ClearDB", "@daily", func() {
			if err := task.ClearTable(global.OPTIMUS_DB); err != nil {
				fmt.Println("timer error:", err)
			}
		}, "clear expired audit logs and revoked tokens", option...)
		if err != nil {
			fmt.Println("add timer error:", err)
		}
	}()
}
