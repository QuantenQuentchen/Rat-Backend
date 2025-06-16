package logic

import (
	"RatBackend/db"
	"fmt"
	"time"
)

func concludeCheck() {
	for {
		time.Sleep(time.Minute)
		voteArr, err := db.GetVotesToConclude()
		if err != nil {
			fmt.Println("There has been an error %w", err)
		}
		for _, vote := range voteArr {
			conclude(vote)
		}
	}
}

func roleRemoveCheck() {
	for {
		time.Sleep(time.Minute)

		roleArr, err := db.GetRolesToRemove()
		if err != nil {
			fmt.Errorf("There has been an error %w", err)
		}
		for _, role := range roleArr {
			err := RemoveRoleBinding(role.User_id, role.Role_id)
			if err != nil {
				fmt.Errorf("There has been an error %w", err)
			}
		}
	}
}

func ScheduledTasks() {
	go concludeCheck()
	go roleRemoveCheck()
}
