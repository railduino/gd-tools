package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"sort"
)

const (
	UserIDsName = "user_ids.json"
)

type User struct {
	Name    string   `json:"name"`
	Comment string   `json:"comment,omitempty"`
	System  bool     `json:"system,omitempty"`
	Shell   string   `json:"shell,omitempty"`
	HomeDir string   `json:"home_dir,omitempty"`
	Groups  []string `json:"groups,omitempty"`
}

type UserID struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
	GID  string `json:"gid"`
}

func (usr *User) String() string {
	return fmt.Sprintf("Name='%s' Gecos='%s' Shell='%s' Home='%s' System=%v",
		usr.Name, usr.Comment, usr.Shell, usr.HomeDir, usr.System)
}

func UsersTest(req *Request) bool {
	return req != nil && len(req.Users) > 0
}

func UsersHandler(req *Request, resp *Response) error {
	if req == nil || len(req.Users) == 0 || resp == nil {
		return nil
	}

	for _, entry := range req.Users {
		if entry.Name == "" {
			return fmt.Errorf("missing user name")
		}
		log.Printf("creating/updating user %s", entry.Name)

		var userCmds []string
		task := "create"
		if _, err := user.Lookup(entry.Name); err != nil {
			addUser := "adduser --group "
			if entry.System {
				addUser += "--system "
			}
			if entry.HomeDir != "" {
				addUser += "--home=" + entry.HomeDir + " "
			}
			addUser += entry.Name
			userCmds = append(userCmds, addUser)
		} else {
			task = "update"
			resp.Sayf("✅ user %s exists", entry.Name)
		}

		if entry.Shell != "" {
			setShell := fmt.Sprintf("usermod --shell %s %s", entry.Shell, entry.Name)
			userCmds = append(userCmds, setShell)
		}

		if entry.Comment != "" {
			setComment := fmt.Sprintf("usermod --comment '%s' %s", entry.Comment, entry.Name)
			userCmds = append(userCmds, setComment)
		}

		if _, err := RunShell(userCmds); err != nil {
			resp.Err = fmt.Sprintf("failed to %s user %s: %v", task, entry.Name, err)
			return err
		}

		if u, err := user.Lookup(entry.Name); err != nil {
			resp.Err = fmt.Sprintf("failed to verify user %s: %v", entry.Name, err)
			return err
		} else {
			userID := UserID{
				Name: u.Username,
				UID:  u.Uid,
				GID:  u.Gid,
			}
			resp.UserIDs = append(resp.UserIDs, userID)
		}

		// TODO (later) check and add groups

		resp.Sayf("✅ user %s was %sd", entry.Name, task)
	}

	return nil
}

func LoadUserIDs() ([]UserID, error) {
	var list []UserID

	content, err := os.ReadFile(UserIDsName)
	if err != nil {
		if os.IsNotExist(err) {
			return list, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", UserIDsName, err)
	}

	if err := json.Unmarshal(content, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", UserIDsName, err)
	}

	return list, nil
}

func GetUserID(name string) (*UserID, error) {
	list, err := LoadUserIDs()
	if err != nil {
		return nil, err
	}

	for _, id := range list {
		if id.Name == name {
			return &id, nil
		}
	}

	return nil, fmt.Errorf("user %s not found", name)
}

func SaveUserIDs(newUser UserID) error {
	oldList, err := LoadUserIDs()
	if err != nil {
		return err
	}

	userMap := make(map[string]UserID)
	for _, u := range oldList {
		userMap[u.Name] = u
	}
	userMap[newUser.Name] = newUser

	var names []string
	for name := range userMap {
		names = append(names, name)
	}
	sort.Strings(names)

	var newList []UserID
	for _, name := range names {
		newList = append(newList, userMap[name])
	}

	content, err := json.MarshalIndent(newList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", UserIDsName, err)
	}
	if err := os.WriteFile(UserIDsName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", UserIDsName, err)
	}

	return nil
}
