<?php
$config['password_driver'] = 'sql';
$config['password_algorithm'] = 'blowfish-crypt';
$config['password_algorithm_prefix'] = '{BLF-CRYPT}';

$config['password_minimum_length'] = 8;
$config['password_require_nonalpha'] = false;

$config['password_db_dsn'] = 'mysql://vmail:{{.Password}}@localhost/vmail';
$config['password_query'] = "UPDATE virtual_users SET user_password = %P WHERE email = %u";

$config['password_confirm_current'] = true;
$config['password_log'] = true;

