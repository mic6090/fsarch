package fsarch

import (
	"bytes"
	"flag"
	toml "github.com/pelletier/go-toml"
	"io"
	"log"
	"os"
	"path"
)

const configName = "fsarch.conf"

type Agent struct {
	Server  string
	Storage string
	Tsfile  string
}

type Server struct {
	Bind     string `default:"0.0.0.0"`
	Port     int    `default:"80"`
	Backup   string
	Datapath string `default:"data"`
	Hashpath string `default:"hash"`
}

type Config struct {
	Agent  Agent
	Server Server
}

func LoadConfig() Config {
	var configPath = flag.String("c", path.Join(GetEtcPath(), configName), "configuration file name")
	flag.Parse()
	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("Config file '%s' open error: %s\n", *configPath, err)
	}
	defer configFile.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, configFile)
	if err != nil {
		log.Fatalf("Config file '%s' read error: %s\n", *configPath, err)
	}
	var config = Config{}
	err = toml.Unmarshal(buf.Bytes(), &config)
	if err != nil {
		log.Fatalf("Config file '%s' parse error: %s\n", *configPath, err)
	}

	config.Server.Datapath = path.Join(config.Server.Backup, config.Server.Datapath)
	config.Server.Hashpath = path.Join(config.Server.Backup, config.Server.Hashpath)

	return config
}
