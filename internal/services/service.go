package services

type UpdateStrategy string

const (
	FlagStrategy    UpdateStrategy = "flag"
	PayloadStrategy UpdateStrategy = "payload"
)

type ACLType string

const (
	ACLArray       ACLType = "array"
	ACLCommaString ACLType = "comma_string"
)

type ServiceConfig struct {
	Name           string
	ResourceGroup  string
	ACLJSONPath    string
	ACLType        ACLType
	UpdateStrategy UpdateStrategy
}

var registry = map[string]ServiceConfig{
	"mongodbflex": {
		Name:           "mongodbflex",
		ResourceGroup:  "instance",
		ACLJSONPath:    "acl.items",
		ACLType:        ACLArray,
		UpdateStrategy: FlagStrategy,
	},
	"postgresflex": {
		Name:           "postgresflex",
		ResourceGroup:  "instance",
		ACLJSONPath:    "network.acl",
		ACLType:        ACLArray,
		UpdateStrategy: FlagStrategy,
	},
	"sqlserverflex": {
		Name:           "sqlserverflex",
		ResourceGroup:  "instance",
		ACLJSONPath:    "network.acl",
		ACLType:        ACLArray,
		UpdateStrategy: FlagStrategy,
	},
	"redis": {
		Name:           "redis",
		ResourceGroup:  "instance",
		ACLJSONPath:    "parameters.sgw_acl",
		ACLType:        ACLCommaString,
		UpdateStrategy: FlagStrategy,
	},
	"valkey": {
		Name:           "valkey",
		ResourceGroup:  "instance",
		ACLJSONPath:    "parameters.sgw_acl",
		ACLType:        ACLCommaString,
		UpdateStrategy: FlagStrategy,
	},
	"opensearch": {
		Name:           "opensearch",
		ResourceGroup:  "instance",
		ACLJSONPath:    "parameters.sgw_acl",
		ACLType:        ACLCommaString,
		UpdateStrategy: FlagStrategy,
	},
	"rabbitmq": {
		Name:           "rabbitmq",
		ResourceGroup:  "instance",
		ACLJSONPath:    "parameters.sgw_acl",
		ACLType:        ACLCommaString,
		UpdateStrategy: FlagStrategy,
	},
	"mariadb": {
		Name:           "mariadb",
		ResourceGroup:  "instance",
		ACLJSONPath:    "parameters.sgw_acl",
		ACLType:        ACLCommaString,
		UpdateStrategy: FlagStrategy,
	},
	"logme": {
		Name:           "logme",
		ResourceGroup:  "instance",
		ACLJSONPath:    "parameters.sgw_acl",
		ACLType:        ACLCommaString,
		UpdateStrategy: FlagStrategy,
	},
	"ske": {
		Name:           "ske",
		ResourceGroup:  "cluster",
		ACLJSONPath:    "extensions.acl.allowedCidrs",
		ACLType:        ACLArray,
		UpdateStrategy: PayloadStrategy,
	},
}

func Get(service, resourceGroup string) (ServiceConfig, bool) {
	cfg, ok := registry[service]
	if !ok {
		return ServiceConfig{}, false
	}
	if cfg.ResourceGroup != resourceGroup {
		return ServiceConfig{}, false
	}
	return cfg, true
}

func Supported() []string {
	services := make([]string, 0, len(registry))
	for _, cfg := range registry {
		services = append(services, cfg.Name+" "+cfg.ResourceGroup)
	}
	return services
}
