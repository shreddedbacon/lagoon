# Auto Idler
This is a service that auto idles development environments
It comprises of two parts
* CLI Idler
* Service Idler

# CLI Idler
The CLI idler checks lagoon development environments for projects that are configured with an `autoIdle` flag set to `1`(enabled)
It performs the following:
* Check OpenShift Project for running `builds`
* Check CLI pod for any non PID1 processes `pgrep -P 0 | tail -n +3 | wc -l | tr -d ' '` (user ssh, drush, etc..)

If either of these conditions are met (running builds, or running processes) then the CLI pod is *not* idled.

# Service Idler
The Service Idler checks lagoon development environments for projects and environments that are configured with an `autoIdle` flag set to `1`(enabled)
It performs the following:
* Check for any hits to the router logs in the last `X` time
* Check OpenShift Project for running `builds`

If there have been no hits, and there are no running builds, then it will:
* Get the deploymentconfig scale for each service
* Patch each service endpoint with all of the other services as `unidle-targets`
* Put the previous scale annotations, then idle the service

# Usage
Once the idler is running, you can trigger by curling the appropriate endpoint.
```
curl --silent --output /dev/null http://auto-idler:3000/idler/cli
curl --silent --output /dev/null http://auto-idler:3000/idler/service
```

# Configuration
Uses all the same configurations from the old auto-idler
Can configure the time since last hits in logs-db using the `SERVICE_IDLER_TIME=4h`
