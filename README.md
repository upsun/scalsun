ScalSun
=========

This tool provide simple auto-scaling on upsun project. 

[[_TOC_]]

## Usage/install
Deploy the **scalsun** binary into your project on build hook.
```
    hook:
        build: |
            mkdir bin
            curl ...
```

Add cron task every minute on `.upsun/config` :
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
#### Syntax
```
Usage of scalsun:
      --min_host_count: int             Minimum host count (default 1)
      --max_host_count int              Maximum host count (default 3)
      --min_cpu_usage_upscale float     Minimum CPU usage in % (for upscale event only) (default 75.0)
      --max_cpu_usage_downscale float   Maximum CPU usage in % (for downscale event only) (default 60.0)
      --min_mem_usage_upscale float     Minimum memory usage in % (for upscale event only) (default 80.0)
      --max_mem_usage_downscale float   Maximum memory usage in % (for downscale event only) (default 20.0)
  -v, --verbose                         Enable verbose mode
  -s, --silent                          Enable silent mode
```
