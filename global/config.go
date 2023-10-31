package global

import (
	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"log"
)

var (
	Port          = 8080
	RedisForQueue = &redis.Options{
		Addr:     "localhost:6379",
		DB:       0,
		Password: "",
		Username: "",
	}
	RedisForPass = &redis.Options{
		Addr:     "localhost:6379",
		DB:       0,
		Password: "",
		Username: "",
	}
	SystemCapacity          = 1000
	MinCapacity             = 50
	ExpireDuration          = 30
	MaxQueuedCapacity int64 = 10
)

func initConfig() {
	viper.SetConfigName("queue")
	viper.AddConfigPath(RootDir + "/config")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	Port = viper.GetInt("port")

	RedisForQueue = &redis.Options{
		Addr:     viper.GetString("redis.queue.addr"),
		DB:       viper.GetInt("redis.queue.db"),
		Username: viper.GetString("redis.queue.username"),
		Password: viper.GetString("redis.queue.password"),
	}

	RedisForPass = &redis.Options{
		Addr:     viper.GetString("redis.pass.addr"),
		DB:       viper.GetInt("redis.pass.db"),
		Username: viper.GetString("redis.pass.username"),
		Password: viper.GetString("redis.pass.password"),
	}

	SystemCapacity = viper.GetInt("systemCapacity")
	MinCapacity = viper.GetInt("minCapacity")
	ExpireDuration = viper.GetInt("expireDuration")
	MaxQueuedCapacity = viper.GetInt64("maxQueueCapacity")

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		err := viper.ReadInConfig()
		if err != nil {
			log.Println(err) //todo logger.log
			return
		}
		//可动态更新配置
		SystemCapacity = viper.GetInt("systemCapacity")
		MinCapacity = viper.GetInt("minCapacity")
		ExpireDuration = viper.GetInt("expireDuration")
		MaxQueuedCapacity = viper.GetInt64("maxQueueCapacity")
	})
}
