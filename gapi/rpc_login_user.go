package gapi

import (
	"context"
	"database/sql"

	db "github.com/tiendat0139/simple-bank/db/sqlc"
	"github.com/tiendat0139/simple-bank/pb"
	"github.com/tiendat0139/simple-bank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", req.Username)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %s", err)
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid password: %s", err)
	}

	accessToken, _, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token: %s", err)
	}

	refresh_token, refreshTokenId, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefresheTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token: %s", err)
	}

	mtdt := server.extractMetadata(ctx)
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshTokenId,
		Username:     user.Username,
		RefreshToken: refresh_token,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
		IsBlocked:    false,
		ExpriresAt:   refreshPayload.ExpiresAt.Time,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %s", err)
	}

	rsp := &pb.LoginUserResponse{
		User: convertUser(user),
		SessionId: session.ID.String(),
		AccessToken: accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiresAt.Time),
		RefreshToken:          refresh_token,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiresAt.Time),
	}

	return rsp, nil
}
