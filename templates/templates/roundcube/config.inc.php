<?php
$config = array();

$config['db_dsnw'] = 'mysql://rc_{{.Name}}:{{.Password}}@localhost/rc_{{.Name}}';

// IMAP
$config['default_host'] = 'ssl://{{.FQDN}}';
$config['default_port'] = 993;

// SMTP
$config['smtp_server'] = 'ssl://{{.FQDN}}';
$config['smtp_port'] = 465;
$config['smtp_user'] = '%u';
$config['smtp_pass'] = '%p';

$config['session_driver'] = 'redis';
$config['session_redis_host'] = '127.0.0.1';
$config['session_redis_port'] = 6379;

$config['support_url'] = 'mailto:{{.SysAdmin}}';
$config['des_key'] = '{{.DesKey}}';
$config['temp_dir'] = '{{.BaseDir}}/temp';
$config['upload_tmp_dir'] = '{{.BaseDir}}/upload';
$config['log_driver'] = 'file';
$config['log_dir'] = '{{.LogsDir}}';

$config['plugins'] = ['managesieve', 'password', 'archive', 'zipdownload'];

$config['managesieve_host'] = '127.0.0.1';
$config['managesieve_port'] = 4190;
$config['managesieve_auth_type'] = null;

$config['skin'] = 'elastic';
$config['product_name'] = '{{.DomainName}} Webmail';
$config['language'] = '{{.Locale}}';

