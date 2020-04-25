package news

import "time"

type News struct {
	ID      int64     `json:"id" xorm:"pk autoincr"`
	Title   string    `json:"title" xorm:"varchar(128) notnull"`
	Created time.Time `json:"created" xorm:"created"`
	Updated time.Time `json:"updated" xorm:"updated"`
}
