<?php
/**
 * Wordpress config file - do not edit
 */

define('DB_NAME',     '{{.DbName}}');
define('DB_USER',     '{{.DbUser}}');
define('DB_PASSWORD', '{{.DbPswd}}');
define('DB_HOST',     'localhost');
define('DB_CHARSET',  'utf8mb4');
define('DB_COLLATE',  'utf8mb4_general_ci');
$table_prefix = 'wp_';

define('AUTH_KEY',         '{{.KeySalt}}_01');
define('SECURE_AUTH_KEY',  '{{.KeySalt}}_02');
define('LOGGED_IN_KEY',    '{{.KeySalt}}_03');
define('NONCE_KEY',        '{{.KeySalt}}_04');
define('AUTH_SALT',        '{{.KeySalt}}_05');
define('SECURE_AUTH_SALT', '{{.KeySalt}}_06');
define('LOGGED_IN_SALT',   '{{.KeySalt}}_07');
define('NONCE_SALT',       '{{.KeySalt}}_08');

define('WPMS_ON',                   true);
define('WPMS_LICENSE_KEY',          '');
define('WPMS_MAILER',               'smtp');
define('WPMS_MAIL_FROM',            '{{.UserEmail}}');
define('WPMS_MAIL_FROM_FORCE',      true);
define('WPMS_MAIL_FROM_NAME',       '{{.UserName}}');
define('WPMS_MAIL_FROM_NAME_FORCE', true);
define('WPMS_SET_RETURN_PATH',      true);
define('WPMS_DO_NOT_SEND',          false);

define('WPMS_SMTP_HOST',            '127.0.0.1');
define('WPMS_SMTP_PORT',            25);
define('WPMS_SSL',                  '');
define('WPMS_SMTP_AUTH',            false);
define('WPMS_SMTP_AUTOTLS',         false);

define('WP_DEBUG',  false);
define('FS_METHOD', 'direct');

if (!defined('ABSPATH')) {
  define('ABSPATH', __DIR__ . '/');
}

require_once ABSPATH . 'wp-settings.php';

