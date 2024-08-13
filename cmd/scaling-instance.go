package main

import (
	"log"
	"os"

	flag "github.com/spf13/pflag"
	entity "github.com/upsun/lib-sun/entity"
	utils "github.com/upsun/lib-sun/utility"
	app "github.com/upsun/scalsun"
	logic "github.com/upsun/scalsun/internal/logic"
)

const (
	APP_NAME = "scalsun"
)

func init() {
	//flag.StringVarP(&app.ArgsM.SrcProvider, "name", "", "", "Apps or Service name")
	flag.IntVarP(&app.ArgsS.HostCountMin, "min_host_count:", "", 1, "Minimum host count")
	flag.IntVarP(&app.ArgsS.HostCountMax, "max_host_count", "", 3, "Maximum host count")
	flag.Float64VarP(&app.ArgsS.CpuUsageMin, "min_cpu_usage_upscale", "", 75.0, "Minimum CPU usage in % (for upscale event only)")
	flag.Float64VarP(&app.ArgsS.CpuUsageMax, "max_cpu_usage_downscale", "", 60.0, "Maximum CPU usage in % (for downscale event only)")
	flag.Float64VarP(&app.ArgsS.MemUsageMin, "min_mem_usage_upscale", "", 80.0, "Minimum memory usage in % (for upscale event only)")
	flag.Float64VarP(&app.ArgsS.MemUsageMax, "max_mem_usage_downscale", "", 20.0, "Maximum memory usage in % (for downscale event only)")

	flag.BoolVarP(&app.Args.Verbose, "verbose", "v", false, "Enable verbose mode")
	flag.BoolVarP(&app.Args.Silent, "silent", "s", false, "Enable silent mode")

	flag.CommandLine.SortFlags = false
	flag.Parse()
}

func main() {
	utils.InitLogger(APP_NAME)
	utils.Disclaimer(APP_NAME)
	utils.StartReporters(APP_NAME)

	projectID := os.Getenv("PLATFORM_PROJECT")
	branch := os.Getenv("PLATFORM_BRANCH")

	if projectID == "" || branch == "" {
		log.Fatal("No PLATFORM_PROJECT and PLATFORM_BRANCH environment variable set!")
	}

	// Init
	projectContext := entity.MakeProjectContext(
		"upsun",
		projectID,
		branch,
	)
	logic.ScalingInstance(projectContext)
}
