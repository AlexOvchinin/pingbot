package model

type User struct {
	ID        int64
	Username  string
	FirstName string
}

func ContainsUser(user *User, users []*User) bool {
	for _, rangeUser := range users {
		if IsSameUser(rangeUser, user) {
			return true
		}
	}

	return false
}

func IsSameUser(left *User, right *User) bool {
	if left == nil && right == nil {
		return true
	}

	if left == nil {
		return false
	}

	if right == nil {
		return false
	}

	return left.ID == right.ID && left.ID != 0 || left.Username == right.Username && left.Username != ""
}

func AddUser(user *User, users []*User) []*User {
	if !ContainsUser(user, users) {
		return append(users, user)
	}

	return users
}

func RemoveUser(user *User, users []*User) []*User {
	result := []*User{}
	for _, rangeUser := range users {
		if !IsSameUser(user, rangeUser) {
			result = append(result, rangeUser)
		}
	}
	return result
}
