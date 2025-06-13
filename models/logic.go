package models

var VoteKindFunctions = map[VoteKind]func(pro, con, total int) bool{
	Simple:    func(pro, con, total int) bool { return pro > con && total%2 != 0 && total > 3 },
	Qualified: func(pro, con, total int) bool { return (pro > (total/3)*2) && pro > con && total%2 != 0 && total > 3 },
	Emergency: func(pro, con, total int) bool { return pro > con },
}

func (p PermFlag) Has(flag PermFlag) bool {
	return p&flag != 0
}
