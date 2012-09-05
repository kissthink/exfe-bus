package main

import (
	"daemon"
	"fmt"
	"github.com/googollee/go-logger"
	"github.com/googollee/go-mysql"
	"github.com/googollee/godis"
	"gobus"
	"os"
)

type Config struct {
	DB struct {
		Addr     string `json:"addr"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		DbName   string `json:"db_name"`
	} `json:"db"`
	Redis struct {
		Netaddr  string `json:"netaddr"`
		Db       int    `json:"db"`
		Password string `json:"password"`
	} `json:"redis"`
	ExfeService struct {
		Addr string `json:"addr"`
		Port int    `json:"port"`
	} `json:"exfe_service"`
	TokenManager struct {
		TableName string `json:"table_name"`
	} `json:"token_manager"`

	loggerOutput *os.File
	loggerFlags  int
}

func main() {
	var config Config
	var quit <-chan os.Signal
	config.loggerOutput, quit = daemon.Init("exfe.json", &config)
	config.loggerFlags = logger.Lshortfile

	log, err := logger.New(config.loggerOutput, "service bus", config.loggerFlags)
	if err != nil {
		panic(err)
	}

	dbAddr := fmt.Sprintf("%s:%d", config.DB.Addr, config.DB.Port)
	db, err := mysql.DialTCP(dbAddr, config.DB.Username, config.DB.Password, config.DB.DbName)
	if err != nil {
		log.Crit("db connect failed: %s", err)
		os.Exit(-1)
	}

	redis := godis.New(config.Redis.Netaddr, config.Redis.Db, config.Redis.Password)

	tkMng, err := NewTokenManager(&config, db, log)
	if err != nil {
		log.Crit("create token manager failed: %s", err)
		os.Exit(-1)
	}

	iom := NewIom(&config, redis, log)

	url := fmt.Sprintf("http://%s:%d", config.ExfeService.Addr, config.ExfeService.Port)
	log.Info("start at %s", url)
	bus, err := gobus.NewGobusServer(url, log)
	if err != nil {
		log.Crit("gobus launch failed: %s", err)
		os.Exit(-1)
	}
	bus.Register(tkMng)
	bus.Register(iom)

	go func() {
		<-quit
		log.Info("quit")
		os.Exit(-1)
	}()
	bus.ListenAndServe()
}