package models

type VoteState int

const (
	Ongoing VoteState = iota
	FailedVeto
	FailedInconclusive
	FailedVotes
	Succeeded
)

type VoteKind int

const (
	Simple VoteKind = iota
	Qualified
	Emergency
)

type Position int

const (
	For Position = iota
	Against
	Abstain
)

type PermFlag uint64

const (
	PermVote         PermFlag = 1 << 0
	PermStartVote    PermFlag = 1 << 1
	PermConcludeVote PermFlag = 1 << 2
	PermVeto         PermFlag = 1 << 3
	PermInternalInfo PermFlag = 1 << 4
	PermPublicInfo   PermFlag = 1 << 5
	PermSuggest      PermFlag = 1 << 6
)

type Role struct {
	Name        string `db:"id" json:"name"`
	Permissions uint64 `db:"perms" json:"permissions"`
	Unique      bool   `db:"unique_role" json:"unique"`
	Timeout     uint64 `db:"timeout" json:"timeout,omitempty"`
	Cascade     bool   `db:"cascade" json:"cascade,omitempty"`
}

type RoleBinding struct {
	User_id        string `db:"user_id"`
	Role_id        string `db:"role_id"`
	Transferal     bool
	Issuer_id      *string `db:"issuer_id"`
	Issuer_Role_id *string `db:"issuer_role_id"`
	IssuedAt       uint64  `db:"issuedAt"`
}

type Vote struct {
	ID          uint64    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Kind        VoteKind  `db:"kind"`
	HasAbstain  bool      `db:"hasAbstain"`
	IsPrivate   bool      `db:"isPrivate"`
	VoteState   VoteState `db:"voteState"`
	Timeout     uint64    `db:"timeout"`
	CreatedAt   uint64    `db:"createdAt"`
}

type User struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

type VoteStateChangeContext struct {
	NewVoteID uint64
	VetoedBy  string
	Reason    string
}

type VoteStateOption func(*VoteStateChangeContext)

func WithNewVoteID(id uint64) VoteStateOption {
	return func(ctx *VoteStateChangeContext) {
		ctx.NewVoteID = id
	}
}

func WithVetoDetails(userID, reason string) VoteStateOption {
	return func(ctx *VoteStateChangeContext) {
		ctx.VetoedBy = userID
		ctx.Reason = reason
	}
}
