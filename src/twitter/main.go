package main

import (
	"twitter/service"
	"twitter/job"
	"config"
	"gobus"
	"gosque"
	"log"
	"flag"
	"fmt"
	"os"
)

const (
	queue_friendship = "twitter:friendship"
	queue_info = "twitter:userinfo"
	queue_tweet = "twitter:tweet"
	queue_dm = "twitter:directmessage"
)

func runService(c *twitter_job.Config) {
	server := gobus.CreateServer(c.Redis.Netaddr, c.Redis.Db, c.Redis.Password, "twitter")

	server.Register(new(twitter_service.FriendshipsExists))
	user := new(twitter_service.Users)
	user.SiteUrl = c.Site_url
	server.Register(user)
	server.Register(new(twitter_service.Statuses))
	d := new(twitter_service.DirectMessages)
	d.SiteUrl = c.Site_url
	server.Register(d)

	go server.Serve(c.Service.Time_out * 1e9)
}

func main() {
	log.SetPrefix("[TwitterSender]")
	log.Printf("Service start")

	var c twitter_job.Config

	var pidfile string
	var configFile string

	flag.StringVar(&pidfile, "pid", "", "Specify the pid file")
	flag.StringVar(&configFile, "config", "twitter.json", "Specify the configuration file")
	flag.Parse()

	config.LoadFile(configFile, &c)

	flag.Parse()
	if pidfile != "" {
		pid, err := os.Create(pidfile)
		if err != nil {
			log.Fatal("Can't create pid(%s): %s", pidfile, err)
			return
		}
		pid.WriteString(fmt.Sprintf("%d", os.Getpid()))
	}

	runService(&c)

	client := gobus.CreateClient(
		c.Redis.Netaddr,
		c.Redis.Db,
		c.Redis.Password,
		"twitter")

	job := twitter_job.Twitter_job{
		Config: &c,
		Client: client,
	}

	queue := gosque.CreateQueue("", 0, "", "twitter")
	err := queue.Register(&job)
	if err != nil {
		log.Fatal(err)
	}
	queue.Serve(c.Service.Time_out * 1e9)

	log.Printf("Service stop")
}