package api

import "github.com/bledniua/rpc/jsonrpc"

const (
	SYSTEM int = iota
	USER
	GROUP
	WORKER
	TASK
	PAYMENT
)

var (
	i int
	//basic
	validateError = func(data error) *jsonrpc.Error {
		return &jsonrpc.Error{
			Code:    SYSTEM,
			Message: "validate error",
			Data:    data.Error(),
		}
	}
	cantGetDataFromDatabase = func(data error) *jsonrpc.Error {
		return &jsonrpc.Error{
			Code:    SYSTEM,
			Message: "cant get data from database",
			Data:    data.Error(),
		}
	}
	cantGetData = &jsonrpc.Error{
		Code:    SYSTEM,
		Message: "cant load data",
	}

	//auth
	loginAlreadyExist = &jsonrpc.Error{
		Code:    USER,
		Message: "login already exist",
	}
	cantCreateHashPassword = &jsonrpc.Error{
		Code:    SYSTEM,
		Message: "cant create hash password",
	}
	cantCreateNewUser = &jsonrpc.Error{
		Code:    SYSTEM,
		Message: "cant create new user",
	}
	cantFindUser = &jsonrpc.Error{
		Code:    USER,
		Message: "cant find user",
	}
	wrongPassword = &jsonrpc.Error{
		Code:    USER,
		Message: "wrong password",
	}
	keyNotFound = &jsonrpc.Error{
		Code:    USER,
		Message: "key not found",
	}
	cantSingIn = &jsonrpc.Error{
		Code:    SYSTEM,
		Message: "cant sing in",
	}
	cantSingOut = &jsonrpc.Error{
		Code:    SYSTEM,
		Message: "cant sing out",
	}
	//auth
	cantCreateNewToken = func(data error) *jsonrpc.Error {
		return &jsonrpc.Error{
			Code:    SYSTEM,
			Message: "cant create new token",
			Data:    data.Error(),
		}
	}
	cantFindToken = func(data error) *jsonrpc.Error {
		return &jsonrpc.Error{
			Code:    USER,
			Message: "cant find token",
			Data:    data.Error(),
		}
	}
	cantRefreshToken = func(data error) *jsonrpc.Error {
		return &jsonrpc.Error{
			Code:    USER,
			Message: "cant refresh token",
			Data:    data.Error(),
		}
	}
	tokenNotFound = jsonrpc.Error{
		Code:    USER,
		Message: "token not found",
	}

	//groups
	cantCreateNewGroup = &jsonrpc.Error{
		Code:    GROUP,
		Message: "cant create new group",
	}
	cantGetGroupList = &jsonrpc.Error{
		Code:    GROUP,
		Message: "cant get group list",
	}
	cantFindGroup = &jsonrpc.Error{
		Code:    GROUP,
		Message: "cant find group",
	}
	cantRemoveWorkersExist = &jsonrpc.Error{
		Code:    GROUP,
		Message: "cant remove group there is some workers exist",
	}
	//Workers
	cantGetWorker = &jsonrpc.Error{
		Code:    WORKER,
		Message: "cant get worker",
	}
	cantGetWorkerID = &jsonrpc.Error{
		Code:    WORKER,
		Message: "cant get worker id",
	}

	cantGetWorkerName = &jsonrpc.Error{
		Code:    WORKER,
		Message: "cant get worker name",
	}

	cantGetWorkersList = &jsonrpc.Error{
		Code:    WORKER,
		Message: "cant get workers list",
	}
	cantSendRequest = &jsonrpc.Error{
		Code:    WORKER,
		Message: "cant send request",
	}
	//Task
	cantCreateNewTask = &jsonrpc.Error{
		Code:    TASK,
		Message: "cant create new task",
	}
	cantGetTaskList = &jsonrpc.Error{
		Code:    TASK,
		Message: "cant get task list",
	}
	//payments
	cantFindPayments = &jsonrpc.Error{
		Code:    PAYMENT,
		Message: "cant find payments",
	}
	cantCreatePayment = func(data error) *jsonrpc.Error {
		return &jsonrpc.Error{
			Code:    PAYMENT,
			Message: "cant create payment",
			Data:    data.Error(),
		}
	}
	cantUpdatePayment = func(data error) *jsonrpc.Error {
		return &jsonrpc.Error{
			Code:    PAYMENT,
			Message: "cant create payment",
			Data:    data.Error(),
		}
	}
)
