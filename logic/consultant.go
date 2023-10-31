package logic

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	" github.com/thinhunan/wonder8/request/queue/global"
	"sync"
	"sync/atomic"
	"time"
)

type consultant struct {
	users           sync.Map
	enteringChannel chan *User
	leavingChannel  chan *User
	rdbForQueue     *redis.Client
	rdbForPass      *redis.Client
	queuedNum       int64
}

var Consultant = &consultant{
	enteringChannel: make(chan *User),
	leavingChannel:  make(chan *User),
}

var ctx = context.Background()

const RedisQueueKey = "q:ips"

func (c *consultant) Start() {
	c.rdbForQueue = redis.NewClient(global.RedisForQueue)

	defer func(rdbForQueue *redis.Client) {
		_ = rdbForQueue.Close()
	}(c.rdbForQueue)

	c.rdbForPass = redis.NewClient(global.RedisForPass)
	defer func(rdbForPass *redis.Client) {
		_ = rdbForPass.Close()
	}(c.rdbForPass)

	go c.CheckAllUser()
	for {
		select {
		case newUser := <-c.enteringChannel:
			r, err := c.rdbForPass.Exists(ctx, newUser.IP).Result()
			if err != nil {
				log.Printf("deal enteringChannel err: %+v", err)
				break
			}
			if r > 0 {
				newUser.MessageChannel <- &Message{
					Left:     0,
					Estimate: 0,
					Position: -1,
					MsgTime:  time.Now(),
					Ip:       newUser.IP,
				}
				newUser.Status = StatusClosing
				break
			}
			if atomic.LoadInt64(&c.queuedNum) >= global.MaxQueuedCapacity {
				newUser.MessageChannel <- &Message{
					Left:     0,
					Estimate: 0,
					Position: -1,
					MsgTime:  time.Now(),
					OverLoad: true,
					Ip:       newUser.IP,
				}
				newUser.Status = StatusClosing
				break
			}
			var (
				u     interface{}
				users map[string]*User
				ok    bool
			)
			if u, ok = c.users.Load(newUser.IP); !ok {
				atomic.AddInt64(&c.queuedNum, 1)
				users = make(map[string]*User)
				users[newUser.Port] = newUser
				c.users.Store(newUser.IP, users)
				_, err = c.rdbForQueue.Do(ctx, "zadd", RedisQueueKey, time.Now().UnixMicro(), newUser.IP).Result()
				if err != nil {
					log.Printf("zadd failed ip: %s err: %+v", newUser.IP, err)
					break
				}
			} else {
				users, _ = u.(map[string]*User)
				if _, ok = users[newUser.Port]; !ok {
					users[newUser.Port] = newUser
				}
			}
		case oldUser := <-c.leavingChannel:
			if u, ok := c.users.Load(oldUser.IP); ok {
				users, _ := u.(map[string]*User)
				delete(users, oldUser.Port)
				if len(users) == 0 {
					atomic.AddInt64(&c.queuedNum, -1)
					c.users.Delete(oldUser.IP)
					_, err := c.rdbForQueue.Do(ctx, "zrem", RedisQueueKey, oldUser.IP).Result()
					if err != nil {
						log.Printf("zrem failed ip: %s err: %+v", oldUser.IP, err)
						break
					}
				}
			} else {
				_, err := c.rdbForQueue.Do(ctx, "zrem", RedisQueueKey, oldUser.IP).Result()
				if err != nil {
					log.Printf("zrem failed ip: %s err: %+v", oldUser.IP, err)
					break
				}
			}
		}
	}
}

func (c *consultant) CheckAllUser() {
	var (
		users map[string]*User
	)
	for {
		passedUserCount, err := c.rdbForPass.DBSize(ctx).Result()
		if err != nil {
			log.Printf("dbsize failed err: %+v\n", err)
			continue
		}
		newPass := global.SystemCapacity - int(passedUserCount)
		if newPass < global.MinCapacity {
			newPass = global.MinCapacity
		}

		luckys, err := c.rdbForQueue.ZPopMin(ctx, RedisQueueKey, int64(newPass)).Result()
		if err != nil {
			log.Printf("zpopmin failed err: %+v\n", err)
			continue
		}
		for _, lucky := range luckys {
			luckyIp := lucky.Member.(string)
			_, err = c.rdbForPass.SetEX(ctx, luckyIp, 1, time.Second*time.Duration(global.ExpireDuration)).Result()
			if err != nil {
				log.Printf("setex failed luckyIp: %s err: %+v\n", luckyIp, err)
				continue
			}
			u, ok := c.users.LoadAndDelete(luckyIp)
			if ok {
				if users, ok = u.(map[string]*User); ok {
					for _, user := range users {
						if user.Status == StatusClosed {
							continue
						}
						user.MessageChannel <- &Message{
							Left:     0,
							Estimate: 0,
							Position: -1,
							MsgTime:  time.Now(),
							Ip:       user.IP,
						}
					}
				}
			}
		}

		queueSize, err := c.rdbForQueue.ZCard(ctx, RedisQueueKey).Result()
		if err != nil {
			log.Printf("zcard failed err: %+v\n", err)
			continue
		}

		if newPass < 1 {
			newPass = 1
		}
		c.users.Range(func(key, value interface{}) bool {
			ip := key.(string)
			rank, err := Consultant.rdbForQueue.ZRank(ctx, RedisQueueKey, ip).Result()
			if err != nil {
				log.Printf("zrank failed ip: %s err: %+v\n", ip, err)
				return true
			}
			users = value.(map[string]*User)
			for _, user := range users {
				if user.Status == StatusClosed {
					continue
				}
				user.MessageChannel <- &Message{
					Left:     int(queueSize),
					Estimate: int(rank)/newPass + 1,
					Position: int(rank) + 1,
					MsgTime:  time.Now(),
					Ip:       user.IP,
				}
			}
			return true
		})
		time.Sleep(time.Second)
	}
}

func (c *consultant) UserEntering(u *User) {
	c.enteringChannel <- u
}

func (c *consultant) UserLeaving(u *User) {
	c.leavingChannel <- u
}
