> [!CAUTION]
> ## This project is owned by the Upsun Advocacy team. It is in an early stage of development [experimental] and only intended to be used by Upsun customers/community.   <br /><br />This project is not supported by Upsun and does not qualify for Support plans. It's provided without guarantee or warranty !

ScalSun
=========

This tool provide simple auto-scaling on upsun project. 

## Usage/install

Deploy the **scalsun** and **Upsun CLI** binary into your project

On `.upsun/config`  
Add to build hook :
```
    hook:
        build: |
            mkdir bin
            curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | VENDOR=upsun bash
            curl -fsSL https://github.com/upsun/scalsun/releases/download/v0.3.3/scalsun-v0.3.3-linux-amd64.tar.gz | tar -xzf - -c bin
```

Add cron task every minute :
```
    crons:
        autoscaller:
            spec: "*/1 * * * *"
            commands:
                start: |
                    if [ "$PLATFORM_ENVIRONMENT_TYPE" = "production" ]; then
                        /app/bin/scalsun --silent --max_host_count=${H_SCALING_HOST_MAX:-3}
                    fi
```

On `Upsun console`,  
Add a environment variables with your [token](https://docs.upsun.com/administration/cli/api-tokens.html#2-create-an-api-token) :
```
env:UPSUN_CLI_TOKEN
```

### Syntax

```
Usage of scalsun:
      --name string                     Apps or Service name
      --min_host_count: int             Minimum host count (default 1)
      --max_host_count int              Maximum host count (default 3)
      --min_cpu_usage_upscale float     Minimum CPU usage in % (for upscale event only) (default 75)
      --max_cpu_usage_downscale float   Maximum CPU usage in % (for downscale event only) (default 60)
      --min_mem_usage_upscale float     Minimum memory usage in % (for upscale event only) (default 80)
      --max_mem_usage_downscale float   Maximum memory usage in % (for downscale event only) (default 20)
  -v, --verbose                         Enable verbose mode
  -s, --silent                          Enable silent mode
      --pathLog string                  Define Path of Log file (default "./")
```

#### Samples
- Auto-scale all app/service  
`scalsun --silent --max_host_count=${H_SCALING_HOST_MAX:-3}`
- Auto-scale only specific app (if app name is web)  
`scalsun --silent --max_host_count=${H_SCALING_HOST_MAX:-3} --name=web`
