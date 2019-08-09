// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

// Config for the rabbitmqbeat
type Config struct {
	RabbitmqHostname    string   `config:"rabbitmq_hostname"`
	RabbitmqPort        string   `config:"rabbitmq_port"`
	RabbitmqUsername    string   `config:"rabbitmq_username"`
	RabbitmqPasswd      string   `config:"rabbitmq_passwd"`
	RabbitmqExchange    string   `config:"rabbitmq_exchange"`
	RabbitmqRoutingKeys []string `config:"rabbitmq_routing_keys"`
	LogConfig           bool     `config:"rabbitmq_log_config"`
}

// DefaultConfig for the rabbitmqbeat
var DefaultConfig = Config{
	RabbitmqHostname:    "localhost",
	RabbitmqPort:        "5672",
	RabbitmqUsername:    "",
	RabbitmqPasswd:      "",
	RabbitmqExchange:    "",
	RabbitmqRoutingKeys: []string{"*.*"},
	LogConfig:           false,
}
