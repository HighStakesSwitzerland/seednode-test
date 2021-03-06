package seednode

import (
  "bytes"
  "fmt"
  "github.com/mitchellh/go-homedir"
  "github.com/spf13/viper"
  "github.com/tendermint/tendermint/config"
  tmos "github.com/tendermint/tendermint/libs/os"
  "github.com/tendermint/tendermint/p2p"

  "os"
  "path/filepath"
  "strings"
  "text/template"
)

// TSConfig extends tendermint P2PConfig with the things we need
type TSConfig struct {
  config.P2PConfig `mapstructure:"p2p"`
  HttpPort         string `mapstructure:"http_port"`
  ChainId          string `mapstructure:"chain_id"`
  LogLevel         string `mapstructure:"log_level"`
}

var configTemplate *template.Template
var configPath = ".seednode-test"

func init() {
  var err error
  tmpl := template.New("configFileTemplate").Funcs(template.FuncMap{
    "StringsJoin": strings.Join,
  })
  if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
    panic(err)
  }
}

func InitConfig() (TSConfig, p2p.NodeKey) {
  var tsConfig TSConfig
  userHomeDir, err := homedir.Dir()
  if err != nil {
    panic(err)
  }

  // init config directory & files if they don't exists yet
  homeDir := filepath.Join(userHomeDir, configPath)
  nodeKeyFilePath := filepath.Join(homeDir, "node_key.json")
  configFilePath := filepath.Join(homeDir, "config.toml")

  if err = os.MkdirAll(homeDir, os.ModePerm); err != nil {
    panic(err)
  }

  viper.SetConfigName("config")
  viper.AddConfigPath(filepath.Join(userHomeDir, configPath))

  if err := viper.ReadInConfig(); err == nil {
    logger.Info(fmt.Sprintf("Loading config file: %s", viper.ConfigFileUsed()))
    err := viper.Unmarshal(&tsConfig)
    if err != nil {
      panic("Invalid config file!")
    }
  } else if _, ok := err.(viper.ConfigFileNotFoundError); ok {
    // ignore not found error, return other errors
    logger.Info("No existing configuration found, generating one")
    tsConfig = initDefaultConfig()
    writeConfigFile(configFilePath, &tsConfig)
  } else {
    panic(err)
  }

  nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFilePath)
  if err != nil {
    panic(err)
  }

  logger.Info("Configuration",
    "path", configFilePath,
    "node listen", tsConfig.ListenAddress,
    "http server port", tsConfig.HttpPort,
    "chain", tsConfig.ChainId,
    "nodeId", nodeKey.ID(),
  )
  if tsConfig.ChainId == "" {
    panic("ChainId is empty")
  }
  return tsConfig, *nodeKey
}

func initDefaultConfig() TSConfig {
  p2PConfig := config.DefaultP2PConfig()
  tsConfig := TSConfig{
    P2PConfig: *p2PConfig,
    HttpPort:  "8090",
    LogLevel:  "info",
  }
  tsConfig.ListenAddress = "tcp://0.0.0.0:44444"
  tsConfig.ChainId = "bombay-12"
  tsConfig.MaxNumInboundPeers = 3000
  tsConfig.MaxNumOutboundPeers = 3000
  tsConfig.PersistentPeersMaxDialPeriod = 5
  return tsConfig
}

// WriteConfigFile renders config using the template and writes it to configFilePath.
func writeConfigFile(configFilePath string, config *TSConfig) {
  var buffer bytes.Buffer

  if err := configTemplate.Execute(&buffer, config); err != nil {
    panic(err)
  }

  tmos.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

# NOTE: Any path below can be absolute (e.g. "/var/myawesomeapp/data") or
# relative to the home directory (e.g. "data"). The home directory is
# "$HOME/.tendermint" by default, but could be changed via $TMHOME env variable
# or --home cmd flag.

#######################################################
###    SeedNode Server Configuration Options      ###
#######################################################
# the ChainId of the network to crawl
chain_id = "{{ .ChainId }}"

http_port = "{{ .HttpPort }}"

# Output level for logging: "info" or "debug". debug will enable pex and addrbook verbose logs
log_level = "{{ .LogLevel }}"

[p2p]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "{{ .ListenAddress }}"
`
