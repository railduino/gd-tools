<?php
define('DB_NAME',     'wp_{{.Name}}');
define('DB_USER',     'wp_{{.Name}}');
define('DB_PASSWORD', '{{.Password}}');
define('DB_HOST',     'localhost');
define('DB_CHARSET',  'utf8mb4');
define('DB_COLLATE',  'utf8mb4_general_ci');
$table_prefix = 'wp_';

define('AUTH_KEY',         '{{.SaltEntry 1}}');
define('SECURE_AUTH_KEY',  '{{.SaltEntry 2}}');
define('LOGGED_IN_KEY',    '{{.SaltEntry 3}}');
define('NONCE_KEY',        '{{.SaltEntry 4}}');
define('AUTH_SALT',        '{{.SaltEntry 5}}');
define('SECURE_AUTH_SALT', '{{.SaltEntry 6}}');
define('LOGGED_IN_SALT',   '{{.SaltEntry 7}}');
define('NONCE_SALT',       '{{.SaltEntry 8}}');

define('WP_DEBUG', false);
// define('WP_DEBUG', true);
// define('WP_DEBUG_DISPLAY', true);
// define('WP_DEBUG_LOG', true);
// @ini_set('display_errors', 1);

define('FS_METHOD', 'direct');
define('WPLANG', '{{.Locale}}');
define('WP_MEMORY_LIMIT', '128M');
define('DISABLE_WP_CRON', true);
define('WP_POST_REVISIONS', 5);
define('AUTOSAVE_INTERVAL', 120);

if (isset($_SERVER['HTTPS']) && $_SERVER['HTTPS'] === 'on') {
    define('FORCE_SSL_ADMIN', true);
}

if (isset($_SERVER['HTTP_X_FORWARDED_PROTO']) && $_SERVER['HTTP_X_FORWARDED_PROTO'] == 'https') {
    $_SERVER['HTTPS'] = 'on';
}

if (!defined('ABSPATH')) {
  define('ABSPATH', __DIR__ . '/');
}

require_once ABSPATH . 'wp-settings.php';

