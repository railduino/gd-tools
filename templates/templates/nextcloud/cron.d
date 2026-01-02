# Cronjob for Nextcloud instance {{.Name}}
# Runs every 5 minutes as www-data

*/5 * * * * www-data test -s {{.BaseDir}}/config/config.php && /usr/bin/php -f {{.BaseDir}}/cron.php

