# Send `telegraf` metrics to Cloudflare Pub/Sub

Depends on https://github.com/influxdata/telegraf/pull/11284

## Install Telegraf

```bash
brew install telegraf
# optional to stop it from ideling in the background
brew services stop 

# generate a config file
telegraf --input-filter cpu:mem:net:swap --output-filter mqtt config > telegraf.conf
```

See [`telegraf.conf`](./telegraf.conf) for an example configuration