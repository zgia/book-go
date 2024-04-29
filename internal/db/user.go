package db

import (
	"fmt"

	"xorm.io/xorm"

	"zgia.net/book/internal/util"
)

type User struct {
	Id         int64  `xorm:"pk autoincr"`
	Username   string `xorm:"unique notnull index"`
	Realname   string `xorm:"notnull"`
	Mobile     string `xorm:"unique notnull"`
	Password   string `xorm:"notnull"`
	Salt       string `xorm:"notnull"`
	Lastip     string
	Lastrealip string
	Createdat  int64 `xorm:"created notnull default 0"`
	Updatedat  int64 `xorm:"updated notnull default 0"`
	Deletedat  int64 `xorm:"deleted notnull default 0"`
}

func (user *User) String() string {
	return fmt.Sprintf("Username: %v, mobile: %v", user.Username, user.Mobile)
}

// CountUsers returns number of users.
func CountUsers() int64 {
	count, _ := x.Count(new(User))
	return count
}

// ListUsers returns number of users in given page.
func ListUsers(page, pageSize int) ([]*User, error) {
	users := make([]*User, 0, pageSize)
	return users, x.Limit(pageSize, (page-1)*pageSize).Asc("journeyid").Find(&users)
}

// InsertUser create a new user without transaction
func InsertUser(username string, sess *xorm.Session) (int64, error) {
	user := &User{
		Username: username,
		Mobile:   username + "@book.zgia.net",
		Password: username,
		Salt:     util.RandStr(4, "cl"),
	}

	return sess.Insert(user)
}

// CreateUser create a new user with transaction
func CreateUser(username string) error {
	var err error

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = InsertUser(username, sess); err != nil {
		return err
	}

	if err = sess.Commit(); err != nil {
		return err
	}

	return nil

}

func GetUser(username string, password string) (*User, error) {
	user := &User{
		Username: username,
	}
	has, err := x.Get(user)

	if err != nil {
		// do nothing
	} else if !has {
		err = fmt.Errorf("user not found: %v", username)
	} else {
		pwd := util.EncodePassword(password, user.Salt)

		if pwd != user.Password {
			err = fmt.Errorf("password not matched: %v", username)
		}
	}

	if err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func ChangeUserPassword(user *User, password string) error {
	var err error

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return err
	}

	u := &User{Password: util.EncodePassword(password, user.Salt)}
	if _, err = sess.ID(user.Id).Cols("password").Update(u); err != nil {
		return err
	}

	if err = sess.Commit(); err != nil {
		return err
	}

	return nil
}
