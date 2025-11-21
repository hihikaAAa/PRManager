package team

import (
	"github.com/hihikaAAa/PRManager/internal/domain/user"
)

type Team struct{
	TeamName string
	Members []*user.User
}