package logic

import (
	"RatBackend/db"
	"fmt"
)

func concludeCheck() {
	voteArr, err := db.GetVotesToConclude()
	if err != nil {
		fmt.Errorf("There has been an error %w", err)
	}
	for _, vote := range voteArr {
		conclude(vote)
	}
}
