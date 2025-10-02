ScalSun
=========

This tool provide simple auto horizontal scaling on Upsun project for your application containers.

> [!CAUTION]
> **This project is owned by the Upsun Advocacy team. It is in early stage of development [experimental] and only intended to be used with caution by Upsun customers/community.   <br /><br />This project is not supported by Upsun and does not qualify for Support plans. Use this repository at your own risks, it is provided without guarantee or warranty!**  
>   
> PS: if you have downloaded that binary and appreciated it, I kindly ask you to let us know on [Discord](https://discord.gg/upsun)  (you can then later discuss with one of our Product manager to share your feedback on it).

## Usage/install

Deploy the **scalsun** and **Upsun CLI** binary into your project

On `.upsun/config`  
Add to build hook :
```
    hook:
        build: |
            curl -fsS https://raw.githubusercontent.com/upsun/snippets/main/src/install-upsun-tool.sh | bash /dev/stdin "scalsun"

```

Add cron task every minute :
```
    crons:
        autoscaler:
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
      --include_service                 Autoscale the services
      --type string                     Type of scaling (horizontal or vertical or timming) (default "horizontal")
      --min_size_count: float           Minimum host size (default 0.1)
      --max_size_count float            Maximum host size (default 8)
      --min_host_count: int             Minimum host count (default 1)
      --max_host_count int              Maximum host count (default 3)
      --min_cpu_usage_upscale float     Minimum CPU usage in % (for upscale event only) (default 75)
      --max_cpu_usage_downscale float   Maximum CPU usage in % (for downscale event only) (default 60)
      --min_mem_usage_upscale float     Minimum memory usage in % (for upscale event only) (default 80)
      --max_mem_usage_downscale float   Maximum memory usage in % (for downscale event only) (default 20)
  -v, --verbose                         Enable verbose mode
  -s, --silent                          Enable silent mode
      --pathLog string                  Define Path of Log file (default "/var/log/")
```

#### Samples
- Auto-scale (Horizontal) all app/service  
`scalsun --silent --max_host_count=${H_SCALING_HOST_MAX:-3}`

- Auto-scale (Horizontal) only specific app (if app name is web)  
`scalsun --silent --type=horizontal --max_host_count=${H_SCALING_HOST_MAX:-3} --name=web`

- Auto-scale (Vertical) only specific app (if app name is web)  
`scalsun --silent --type=vertical --max_cpu_usage_downscale=20 --max_host_size=1 --min_host_size=0.1 --name=web`

- Auto-scale (Custom) only specific app (if app name is web)  
`scalsun --silent --type=custom --max_host_size=1 --max_host_count=0.1 --name=web`
