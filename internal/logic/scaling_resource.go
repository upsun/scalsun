package logic

import (
	"fmt"
	"log"
	"math"
	"slices"
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
	Name   string
	Values []UsageValue

	// Horizontal scaling
	InstanceUpdate bool
	InstanceOld    int
	InstanceNew    int

	// Vertical scaling
	IdProfileUpdate bool
	IdProfileOld    float64
	IdProfileNew    float64
}

type SizeProfile struct {
	Id  string
	Cpu float64
	Mem int32
}

func ForceContainerInstance(projectContext entity.ProjectGlobal) {
	usages := map[string]UsageApp{}

	log.Println("Process Force Scaling... (only use max_host_count & max_size_count argument)")

	sizesAvailable := ListContainerSize(projectContext)
	if slices.Contains(sizesAvailable, app.ArgsS.HostSizeMax) {

	}

	GetContainersUsage(projectContext, usages)

	for _, usage := range usages {
		if usage.IdProfileOld < app.ArgsS.HostSizeMax {
			usage.IdProfileNew = app.ArgsS.HostSizeMax
			usage.IdProfileUpdate = true
		}

		if usage.InstanceOld < app.ArgsS.HostCountMax {
			usage.InstanceNew = app.ArgsS.HostCountMax
			usage.InstanceUpdate = true
		}

		usages[usage.Name] = usage
	}

	SetResources(projectContext, usages)
}

func ScalingContainer(projectContext entity.ProjectGlobal) {
	usages := map[string]UsageApp{}

	log.Println("Process Vertical Auto-Scaling...")

	GetProjectMetrics(projectContext, usages)
	GetContainersUsage(projectContext, usages)
	RemoveUnscaleContainers(projectContext, usages)

	// Compute
	log.Println("Compute Horizontal trend...")
	for _, usage := range usages {

		//// Algo Threshold-limit
		lastMetric := usage.Values[len(usage.Values)-1]

		// CPU case
		if lastMetric.Cpu > app.ArgsS.CpuUsageMin && usage.IdProfileOld < app.ArgsS.HostSizeMax {
			idProfilProposal := float64(math.Ceil(float64(usage.IdProfileOld) * (lastMetric.Cpu / app.ArgsS.CpuUsageMin)))
			if usage.IdProfileNew < idProfilProposal {
				usage.IdProfileNew = idProfilProposal
				if int(usage.IdProfileNew) != int(usage.IdProfileOld) {
					log.Printf("Upscale instance %v from %v to %v ! (Cpu: %v > %v)", usage.Name, usage.IdProfileOld, usage.IdProfileNew, lastMetric.Cpu, app.ArgsS.CpuUsageMin)
					usage.IdProfileUpdate = true
				}
			}
			usages[usage.Name] = usage
		} else if lastMetric.Cpu < app.ArgsS.CpuUsageMax && usage.IdProfileOld > app.ArgsS.HostSizeMin {
			idProfilProposal := float64(math.Ceil(float64(usage.IdProfileOld) * (lastMetric.Cpu / app.ArgsS.CpuUsageMax)))
			if usage.IdProfileNew < idProfilProposal {
				usage.IdProfileNew = idProfilProposal
				if usage.IdProfileNew != usage.IdProfileOld {
					log.Printf("Downscale instance %v from %v to %v ! (Cpu: %v < %v)", usage.Name, usage.IdProfileOld, usage.IdProfileNew, lastMetric.Cpu, app.ArgsS.CpuUsageMax)
					usage.IdProfileUpdate = true
				}
			}
			usages[usage.Name] = usage
		}

		// Mem case
		if lastMetric.Mem > app.ArgsS.MemUsageMin && usage.IdProfileOld < app.ArgsS.HostSizeMax {
			idProfilProposal := float64(math.Ceil(float64(usage.IdProfileOld) * (lastMetric.Mem / app.ArgsS.MemUsageMin)))
			if usage.IdProfileNew < idProfilProposal {
				usage.IdProfileNew = idProfilProposal
				if usage.IdProfileNew != usage.IdProfileOld {
					log.Printf("Upscale instance %v from %v to %v ! (Mem: %v > %v)", usage.Name, usage.IdProfileOld, usage.IdProfileNew, lastMetric.Mem, app.ArgsS.MemUsageMin)
					usage.InstanceUpdate = true
				}
			}
			usages[usage.Name] = usage
		} else if lastMetric.Mem < app.ArgsS.MemUsageMax && usage.IdProfileOld > app.ArgsS.HostSizeMin {
			idProfilProposal := float64(math.Ceil(float64(usage.IdProfileOld) * (lastMetric.Mem / app.ArgsS.MemUsageMax)))
			if usage.IdProfileNew < idProfilProposal {
				usage.IdProfileNew = idProfilProposal
				if usage.IdProfileNew != usage.IdProfileOld {
					log.Printf("Downscale instance %v from %v to %v ! (Mem: %v > %v)", usage.Name, usage.IdProfileOld, usage.IdProfileNew, lastMetric.Mem, app.ArgsS.MemUsageMax)
					usage.InstanceUpdate = true
				}
			}
			usages[usage.Name] = usage
		}
	}

	SetResources(projectContext, usages)
}

func ScalingInstance(projectContext entity.ProjectGlobal) {
	usages := map[string]UsageApp{}

	log.Println("Process Horizontal Auto-Scaling...")

	GetProjectMetrics(projectContext, usages)
	GetContainersUsage(projectContext, usages)
	RemoveUnscaleContainers(projectContext, usages)

	// Compute
	log.Println("Compute Horizontal trend...")
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
					usage.InstanceUpdate = true
				}
			}
			usages[usage.Name] = usage
		} else if lastMetric.Cpu < app.ArgsS.CpuUsageMax && usage.InstanceOld > app.ArgsS.HostCountMin {
			instanceProposal := int(math.Ceil(float64(usage.InstanceOld) * (lastMetric.Cpu / app.ArgsS.CpuUsageMax)))
			if usage.InstanceNew < instanceProposal {
				usage.InstanceNew = instanceProposal
				if usage.InstanceNew != usage.InstanceOld {
					log.Printf("Downscale instance %v from %v to %v ! (Cpu: %v < %v)", usage.Name, usage.InstanceOld, usage.InstanceNew, lastMetric.Cpu, app.ArgsS.CpuUsageMax)
					usage.InstanceUpdate = true
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
					usage.InstanceUpdate = true
				}
			}
			usages[usage.Name] = usage
		} else if lastMetric.Mem < app.ArgsS.MemUsageMax && usage.InstanceOld > app.ArgsS.HostCountMin {
			instanceProposal := int(math.Ceil(float64(usage.InstanceOld) * (lastMetric.Mem / app.ArgsS.MemUsageMax)))
			if usage.InstanceNew < instanceProposal {
				usage.InstanceNew = instanceProposal
				if usage.InstanceNew != usage.InstanceOld {
					log.Printf("Downscale instance %v from %v to %v ! (Mem: %v > %v)", usage.Name, usage.InstanceOld, usage.InstanceNew, lastMetric.Mem, app.ArgsS.MemUsageMax)
					usage.InstanceUpdate = true
				}
			}
			usages[usage.Name] = usage
		}
	}

	SetResources(projectContext, usages)
}

// ListContainerSize retrieves the size profiles of containers in the project.
func ListContainerSize(projectContext entity.ProjectGlobal) []float64 { // SizeProfile {
	log.Println("Get Size of Profile available...")
	payload := []string{
		"--environment=" + projectContext.DefaultEnv,
		"--service=" + app.ArgsS.Name,
		"--no-header",
		"--format=csv",
		"--no-interaction",
		"--yes",
	}
	output, err := utils.CallCLIString(projectContext, "resources:size:list", payload...)
	if err != nil {
		log.Fatalf("command execution failed: %s", err)
	}

	// Parse output
	size_tpl := strings.Split(output, "\n")
	var sizeProfiles []float64 // SizeProfile
	for _, size_str := range size_tpl[:len(size_tpl)-1] {
		size := strings.Split(size_str, ",")

		// Normalize
		// cpu, _ := strconv.ParseFloat(size[1], 64)
		// mem, _ := strconv.ParseInt(size[2], 10, 32)
		// profile := SizeProfile{
		// 	Id:  size[0],
		// 	Cpu: cpu,
		// 	Mem: int32(mem),
		// }
		sizeId, _ := strconv.ParseFloat(size[0], 64)
		sizeProfiles = append(sizeProfiles, sizeId) // profile)
	}

	return sizeProfiles
}

func GetProjectMetrics(projectContext entity.ProjectGlobal, usages map[string]UsageApp) {
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
}

func GetContainersUsage(projectContext entity.ProjectGlobal, usages map[string]UsageApp) {
	log.Println("Get size of containers and number of instance of project...")

	payload := []string{
		"--environment=" + projectContext.DefaultEnv,
		"--format=csv",
		"--columns=service,instance_count,cpu",
		"--no-header",
		"--no-interaction",
		"--yes",
	}
	output, err := utils.CallCLIString(projectContext, "resources:get", payload...)
	if err != nil {
		log.Fatalf("command execution failed: %s", err)
	}

	// Parse output
	apps := strings.Split(output, "\n")
	for _, app_str := range apps[:len(apps)-1] {
		app_raw := strings.Split(app_str, ",")

		// Normalize
		name := app_raw[0]
		idProfil, _ := strconv.ParseFloat(app_raw[2], 64)
		instance, err := strconv.Atoi(app_raw[1])
		if err != nil {
			break
		}

		// Get or create app/service usage
		if app.ArgsS.Name == "" || app.ArgsS.Name == name {
			usage, found := usages[name]
			if !found {
				usage = UsageApp{Name: name}
			}

			// Assign
			usage.InstanceOld = instance
			usage.IdProfileOld = idProfil
			usages[name] = usage
		}
	}
}

func RemoveUnscaleContainers(projectContext entity.ProjectGlobal, usages map[string]UsageApp) {

	if !app.ArgsS.IncludeServices {
		// Get Services
		log.Println("Dectect available services of project... (to exclude)")
		payload := []string{
			"--environment=" + projectContext.DefaultEnv,
			"--format=csv",
			"--columns=name",
			"--no-header",
			"--no-interaction",
			"--yes",
		}
		output, err := utils.CallCLIString(projectContext, "service:list", payload...)
		if err != nil {
			log.Fatalf("command execution failed: %s", err)
		}

		// Parse output
		srvs := strings.Split(output, "\n")
		for _, srv := range srvs[:len(srvs)-1] {
			delete(usages, srv)
		}
	}
}

func SetResources(projectContext entity.ProjectGlobal, usages map[string]UsageApp) {
	log.Println("Set new number of instance...")

	var updateInstance string
	var updateContainer string
	for _, usage := range usages {
		if usage.InstanceUpdate {
			if updateInstance != "" {
				updateInstance += ","
			}
			updateInstance += usage.Name
			updateInstance += ":"
			updateInstance += strconv.Itoa(usage.InstanceNew)
		}
	}

	log.Println("Set new size of containers...")
	for _, usage := range usages {
		if usage.IdProfileUpdate {
			if updateContainer != "" {
				updateContainer += ","
			}
			updateContainer += usage.Name
			updateContainer += ":"
			updateContainer += fmt.Sprintf("%.1f", usage.IdProfileNew)
		}
	}

	if updateInstance != "" || updateContainer != "" {
		log.Println("Apply new resources...")
		payload := []string{
			"--environment=" + projectContext.DefaultEnv,
			"--no-interaction",
			"--yes",
			"--no-wait",
		}
		if updateInstance != "" {
			payload = append(payload, "--count", updateInstance)
		}
		if updateContainer != "" {
			payload = append(payload, "--size", updateContainer)
		}

		output, err := utils.CallCLIString(projectContext, "resources:set", payload...)
		if err != nil {
			log.Fatalf("command execution failed: %s", err)
		}
		fmt.Print(output)
	} else {
		log.Println("Nothing to do !")
	}
}
