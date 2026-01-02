<?php
define('DB_TYPE', 'mysql');
define('DB_HOST', 'localhost');
define('DB_NAME', 'wbce_{{.Name}}');
define('DB_USERNAME', 'wbce_{{.Name}}');
define('DB_PASSWORD', '{{.Password}}');

define('TABLE_PREFIX', 'wbce_');
define('WB_PATH', dirname(__FILE__));
define('WB_URL', 'https://{{.FQDN}}');

define('DEBUG', false);
define('DEBUG_ADMIN', false);

