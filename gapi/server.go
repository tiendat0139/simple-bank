package gapi

import (
	"fmt"

	db "github.com/tiendat0139/simple-bank/db/sqlc"
	"github.com/tiendat0139/simple-bank/token"
	"github.com/tiendat0139/simple-bank/util"
	"github.com/tiendat0139/simple-bank/pb"
)


type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
