package logic

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	entity "github.com/upsun/lib-sun/entity"
	utils "github.com/upsun/lib-sun/utility"
	app "github.com/upsun/scalsun"
)

type UsageValue struct {
	Cpu float64
	Mem float64
}

type UsageApp struct {
	Name        string
	Values      []UsageValue
	Update      bool
	InstanceOld int
	InstanceNew int
}

func ScalingInstance(projectContext entity.ProjectGlobal) {
	usages := map[string]UsageApp{}

	// Get metrics
	log.Println("Get metrics of project...")
	payload := []string{
		"--environment=" + projectContext.DefaultEnv,
		"--interval=1m",
		"--format=csv",
		"--columns=service,cpu_percent,mem_percent",
		"--no-header",
		"--no-interaction",
		"--yes",
	}
	output, err := utils.CallCLIString(projectContext, "metrics:all", payload...)
	if err != nil {
		log.Fatalf("command execution failed: %s", err)
	}

	// Parse output
	apps := strings.Split(output, "\n")
	for _, app_str := range apps[:len(apps)-1] {
		app := strings.Split(app_str, ",")

		// Normalize
		name := app[0]
		cpu, _ := strconv.ParseFloat(strings.Replace(app[1], "%", "", 1), 64)
		mem, _ := strconv.ParseFloat(strings.Replace(app[2], "%", "", 1), 64)

		// Get or create app/service usage
		usage, found := usages[name]
		if !found {
			usage = UsageApp{Name: name}
		}

		// Assign
		value := UsageValue{Cpu: cpu, Mem: mem}
		usage.Values = append(usage.Values, value)
		usages[name] = usage
	}

	// Get instance
	log.Println("Get number of instance of project...")
	payload = []string{
		"--environment=" + projectContext.DefaultEnv,
		"--format=csv",
		"--columns=service,instance_count",
		"--no-header",
		"--no-interaction",
		"--yes",
	}
	output, err = utils.CallCLIString(projectContext, "resources:get", payload...)
	if err != nil {
		log.Fatalf("command execution failed: %s", err)
	}

	// Parse output
	apps = strings.Split(output, "\n")
	for _, app_str := range apps[:len(apps)-1] {
		app := strings.Split(app_str, ",")

		// Normalize
		name := app[0]
		instance, err := strconv.Atoi(app[1])
		if err != nil {
			break
		}

		// Get or create app/service usage
		usage, found := usages[name]
		if !found {
			usage = UsageApp{Name: name}
		}

		// Assign
		usage.InstanceOld = instance
		usages[name] = usage
	}

	if !app.ArgsS.IncludeServices {
		// Get Services
		log.Println("Dectect available services of project... (to exclude)")
		payload = []string{
			"--environment=" + projectContext.DefaultEnv,
			"--format=csv",
			"--columns=name",
			"--no-header",
			"--no-interaction",
			"--yes",
		}
		output, err = utils.CallCLIString(projectContext, "service:list", payload...)
		if err != nil {
			log.Fatalf("command execution failed: %s", err)
		}

		// Parse output
		srvs := strings.Split(output, "\n")
		for _, srv := range srvs[:len(srvs)-1] {
			delete(usages, srv)
		}
	}

	// Compute
	log.Println("Compute trend...")

	for _, usage := range usages {

		//// Algo Threshold-limit
		lastMetric := usage.Values[len(usage.Values)-1]

		// CPU case
		if lastMetric.Cpu > app.ArgsS.CpuUsageMin && usage.InstanceOld < app.ArgsS.HostCountMax {
			instanceProposal := int(math.Ceil(float64(usage.InstanceOld) * (lastMetric.Cpu / app.ArgsS.CpuUsageMin)))
			if usage.InstanceNew < instanceProposal {
				usage.InstanceNew = instanceProposal
				if usage.InstanceNew != usage.InstanceOld {
					log.Printf("Upscale instance %v from %v to %v ! (Cpu: %v > %v)", usage.Name, usage.InstanceOld, usage.InstanceNew, lastMetric.Cpu, app.ArgsS.CpuUsageMin)
					usage.Update = true
				}
			}
			usages[usage.Name] = usage
		} else if lastMetric.Cpu < app.ArgsS.CpuUsageMax && usage.InstanceOld > app.ArgsS.HostCountMin {
			instanceProposal := int(math.Ceil(float64(usage.InstanceOld) * (lastMetric.Cpu / app.ArgsS.CpuUsageMax)))
			if usage.InstanceNew < instanceProposal {
				usage.InstanceNew = instanceProposal
				if usage.InstanceNew != usage.InstanceOld {
					log.Printf("Downscale instance %v from %v to %v ! (Cpu: %v < %v)", usage.Name, usage.InstanceOld, usage.InstanceNew, lastMetric.Cpu, app.ArgsS.CpuUsageMax)
					usage.Update = true
				}
			}
			usages[usage.Name] = usage
		}

		// Mem case
		if lastMetric.Mem > app.ArgsS.MemUsageMin && usage.InstanceOld < app.ArgsS.HostCountMax {
			instanceProposal := int(math.Ceil(float64(usage.InstanceOld) * (lastMetric.Mem / app.ArgsS.MemUsageMin)))
			if usage.InstanceNew < instanceProposal {
				usage.InstanceNew = instanceProposal
				if usage.InstanceNew != usage.InstanceOld {
					log.Printf("Upscale instance %v from %v to %v ! (Mem: %v > %v)", usage.Name, usage.InstanceOld, usage.InstanceNew, lastMetric.Mem, app.ArgsS.MemUsageMin)
					usage.Update = true
				}
			}
			usages[usage.Name] = usage
		} else if lastMetric.Mem < app.ArgsS.MemUsageMax && usage.InstanceOld > app.ArgsS.HostCountMin {
			instanceProposal := int(math.Ceil(float64(usage.InstanceOld) * (lastMetric.Mem / app.ArgsS.MemUsageMax)))
			if usage.InstanceNew < instanceProposal {
				usage.InstanceNew = instanceProposal
				if usage.InstanceNew != usage.InstanceOld {
					log.Printf("Downscale instance %v from %v to %v ! (Mem: %v > %v)", usage.Name, usage.InstanceOld, usage.InstanceNew, lastMetric.Mem, app.ArgsS.MemUsageMax)
					usage.Update = true
				}
			}
			usages[usage.Name] = usage
		}
	}

	// Reassign resource
	log.Println("Set new number of instance...")
	var update string
	for _, usage := range usages {
		if usage.Update {
			if update != "" {
				update += ","
			}
			update += usage.Name
			update += ":"
			update += strconv.Itoa(usage.InstanceNew)
		}
	}

	if update != "" {
		payload = []string{
			"--environment=" + projectContext.DefaultEnv,
			"--no-interaction",
			"--yes",
			"--no-wait",
			"--count",
			update,
		}
		output, err = utils.CallCLIString(projectContext, "resources:set", payload...)
		if err != nil {
			log.Fatalf("command execution failed: %s", err)
		}
		fmt.Print(output)
	} else {
		log.Println("Nothing to do !")
	}
}
