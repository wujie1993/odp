[defaults]
roles_path = {{ . }}/roles
library = {{ . }}/library
host_key_checking = false
strategy = mitogen_linear
strategy_plugins = /usr/lib/python2.7/site-packages/ansible_mitogen/plugins/strategy
callback_whitelist = profile_tasks, dense
fact_caching=jsonfile
fact_caching_connection = /tmp/ansible
stdout_callback = yaml
