package api

import "RatBackend/models"

func ResolveCall(function, args string) (interface{}, error) {
	switch function {
	case "CreateUser":
		if req, ok := args.(CreateUserRequest); ok {
			return nil, CreateUser(req)
		}
	case "AssignRole":
		if req, ok := args.(AssignRoleRequest); ok {
			return nil, AssignRole(req)
		}
	case "CreateVote":
		if req, ok := args.(VoteCreationRequest); ok {
			return nil, CreateVote(req)
		}
	case "VetoVote":
		if req, ok := args.(VoteVetoRequest); ok {
			return nil, VetoVote(req)
		}
	case "SetVotePrivate":
		if req, ok := args.(VotePrivateRequest); ok {
			return nil, SetVotePrivate(req)
		}
	case "CreateRole":
		if req, ok := args.(CreateRoleRequest); ok {
			_, err := CreateRole(req)
			return nil, err
		}
	case "ModifyRole":
		if req, ok := args.(ModifyRoleRequest); ok {
			return nil, ModifyRole(req)
		}

	default:
		return nil, models.ErrUnknownFunction
	}
	return nil, nil
}
