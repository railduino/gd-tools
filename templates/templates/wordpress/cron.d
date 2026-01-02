# Cronjob for WordPress instance {{.Name}}
# Runs every 5 minutes as www-data

*/5 * * * * root test -x {{.WP_CLI_Path}} && {{.WP_CLI_Path}} --cron >>{{.LogsDir}}/cron.log 2>&1

