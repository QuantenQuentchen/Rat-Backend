package models

type VoteState int

const (
	Ongoing VoteState = iota
	FailedVeto
	FailedInconclusive
	FailedVotes
	Succeeded
	SucceededEmergency
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
	PermCreateRole   PermFlag = 1 << 7
	PermAuditLog     PermFlag = 1 << 8
	PermContestVeto  PermFlag = 1 << 9
	PermSuspendUser  PermFlag = 1 << 10
	PermRemoveRole   PermFlag = 1 << 11
	PermAssignRole   PermFlag = 1 << 12
	PermRestoreUser  PermFlag = 1 << 13
)

var PermFlagNames = map[PermFlag]string{
	PermVote:         "vote",
	PermStartVote:    "start_vote",
	PermConcludeVote: "conclude_vote",
	PermVeto:         "veto",
	PermInternalInfo: "internal_info",
	PermPublicInfo:   "public_info",
	PermSuggest:      "suggest",
	PermCreateRole:   "create_role",
	PermAuditLog:     "audit_log",
	PermContestVeto:  "contest_veto",
	PermSuspendUser:  "suspend_user",
	PermRemoveRole:   "remove_role",
	PermAssignRole:   "assign_role",
	PermRestoreUser:  "restore_user",
}

type Role struct {
	ID          uint64 `db:"id" json:"id, omitempty"`
	Name        string `db:"name" json:"name"`
	Permissions uint64 `db:"perms" json:"permissions"`
	Unique      bool   `db:"unique_role" json:"unique"`
	Timeout     uint64 `db:"timeout" json:"timeout,omitempty"`
	Cascade     bool   `db:"cascade" json:"cascade,omitempty"`
}

type RoleBinding struct {
	UserId       string `db:"user_id"`
	RoleId       string `db:"role_id"`
	Transferal   bool
	IssuerId     *string `db:"issuer_id"`
	IssuerRoleId *string `db:"issuer_role_id"`
	IssuedAt     uint64  `db:"issuedAt"`
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

type RoleChangeReason int

const (
	ReasonNone RoleChangeReason = iota
	ReasonTransferal
	ReasonTimeout
	ReasonCascading
	ReasonRemovedByUser
)

type RoleChangeContext struct {
	Reason      RoleChangeReason
	RemovedByID string
	IsCascading bool
}

type RoleChangeOption func(*RoleChangeContext)

func WithRoleChangeReason(reason RoleChangeReason) RoleChangeOption {
	return func(ctx *RoleChangeContext) {
		ctx.Reason = reason
	}
}

func WithCascadeReason() RoleChangeOption {
	return func(ctx *RoleChangeContext) {
		ctx.IsCascading = true
	}
}

func WithRemovedByID(userID string) RoleChangeOption {
	return func(ctx *RoleChangeContext) {
		ctx.Reason = ReasonRemovedByUser
		ctx.RemovedByID = userID
	}
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
