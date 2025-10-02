package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
	lib "github.com/upsun/lib-sun"
	entity "github.com/upsun/lib-sun/entity"
	utils "github.com/upsun/lib-sun/utility"
	app "github.com/upsun/scalsun"
	logic "github.com/upsun/scalsun/internal/logic"
)

const (
	APP_NAME        = "scalsun"
	TYPE_HORIZONTAL = "horizontal"
	TYPE_VERTICAL   = "vertical"
	TYPE_CUSTOM     = "custom"
)

func init() {
	// Common arguments
	flag.StringVarP(&app.ArgsS.Name, "name", "", "", "Apps or Service name")
	flag.BoolVarP(&app.ArgsS.IncludeServices, "include_service", "", false, "Autoscale the services (not applicable)")

	flag.StringVarP(&app.ArgsS.Type, "type", "", TYPE_HORIZONTAL, fmt.Sprintf("Type of scaling (%s or %s or %s)", TYPE_HORIZONTAL, TYPE_VERTICAL, TYPE_CUSTOM))

	// Vertical scaling arguments
	flag.Float64VarP(&app.ArgsS.HostSizeMin, "min_size_count", "", 0.1, "Minimum host size")
	flag.Float64VarP(&app.ArgsS.HostSizeMax, "max_size_count", "", 8, "Maximum host size")

	// Horizontal scaling arguments
	flag.IntVarP(&app.ArgsS.HostCountMin, "min_host_count", "", 1, "Minimum host count")
	flag.IntVarP(&app.ArgsS.HostCountMax, "max_host_count", "", 3, "Maximum host count")

	// Trigger
	//flag.StringVarP(&app.ArgsS.RangeTime, "", "", "", "")
	flag.Float64VarP(&app.ArgsS.CpuUsageMin, "min_cpu_usage_upscale", "", 75.0, "Minimum CPU usage in % (for upscale event only)")
	flag.Float64VarP(&app.ArgsS.CpuUsageMax, "max_cpu_usage_downscale", "", 60.0, "Maximum CPU usage in % (for downscale event only)")
	flag.Float64VarP(&app.ArgsS.MemUsageMin, "min_mem_usage_upscale", "", 80.0, "Minimum memory usage in % (for upscale event only)")
	flag.Float64VarP(&app.ArgsS.MemUsageMax, "max_mem_usage_downscale", "", 20.0, "Maximum memory usage in % (for downscale event only)")

	flag.BoolVarP(&app.Args.Verbose, "verbose", "v", false, "Enable verbose mode")
	flag.BoolVarP(&app.Args.Silent, "silent", "s", false, "Enable silent mode")
	flag.StringVarP(&app.Args.PathLog, "pathLog", "", "/var/log/", "Define Path of Log file")

	flag.CommandLine.SortFlags = false
	flag.Parse()

	//TODO(Mick) Hack (replace by context object)
	lib.VERSION = app.VERSION
	lib.Args = app.Args
}

func main() {
	utils.InitLogger(APP_NAME)
	utils.Disclaimer(APP_NAME)
	utils.StartReporters(APP_NAME)

	projectID := os.Getenv("PLATFORM_PROJECT")
	branch := os.Getenv("PLATFORM_BRANCH")

	if projectID == "" || branch == "" {
		log.Fatal("No PLATFORM_PROJECT and PLATFORM_BRANCH environment variable set!")
	} else {
		//TODO(Mick) Hack (replace by context object)
		app.Args.PathLog = "/var/log"
		lib.Args = app.Args
	}

	if !entity.IsValidType(app.ArgsS.Type) {
		fmt.Fprintf(os.Stderr, "Invalid value for --type: %q\n", app.ArgsS.Type)
		fmt.Fprintf(os.Stderr, "Valid values are: %s\n", strings.Join(entity.ValidTypes, ", "))
		os.Exit(1)
	}

	// Init
	projectContext := entity.MakeProjectContext(
		"upsun",
		projectID,
		branch,
	)

	switch app.ArgsS.Type {
	case TYPE_CUSTOM:
		logic.ForceContainerInstance(projectContext)
	case TYPE_VERTICAL:
		logic.ScalingContainer(projectContext)
	case TYPE_HORIZONTAL:
		logic.ScalingInstance(projectContext)
	}
}
