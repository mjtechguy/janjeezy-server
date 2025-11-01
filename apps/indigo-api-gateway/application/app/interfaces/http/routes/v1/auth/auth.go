package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"menlo.ai/indigo-api-gateway/app/domain/auth"
	"menlo.ai/indigo-api-gateway/app/domain/user"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/responses"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/auth/google"
	"menlo.ai/indigo-api-gateway/app/utils/idgen"
)

type AuthRoute struct {
	google      *google.GoogleAuthAPI
	userService *user.UserService
	authService *auth.AuthService
}

func NewAuthRoute(
	google *google.GoogleAuthAPI,
	userService *user.UserService,
	authService *auth.AuthService) *AuthRoute {
	return &AuthRoute{
		google,
		userService,
		authService,
	}
}

func (authRoute *AuthRoute) RegisterRouter(router gin.IRouter) {
	authRouter := router.Group("/auth")
	authRouter.GET("/logout", authRoute.Logout)
	authRouter.GET("/refresh-token", authRoute.RefreshToken)
	authRouter.GET("/me",
		authRoute.authService.AppUserAuthMiddleware(),
		authRoute.authService.RegisteredUserMiddleware(),
		authRoute.GetMe,
	)
	authRouter.POST("/guest-login", authRoute.GuestLogin)
	authRouter.POST("/local/login", authRoute.LocalLogin)
	authRoute.google.RegisterRouter(authRouter)

}

// @Enum(access.token)
type AccessTokenResponseObjectType string

const AccessTokenResponseObjectTypeObject = "access.token"

type AccessTokenResponse struct {
	Object      AccessTokenResponseObjectType `json:"object"`
	AccessToken string                        `json:"access_token"`
	ExpiresIn   int                           `json:"expires_in"`
}

type LocalLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type GetMeResponse struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// @Summary Get user profile
// @Description Retrieves the profile of the authenticated user based on the provided JWT.
// @Tags Authentication API
// @Security BearerAuth
// @Produce json
// @Success 200 {object} GetMeResponse "Successfully retrieved user profile"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized (e.g., missing or invalid JWT)"
// @Router /v1/auth/me [get]
func (authRoute *AuthRoute) GetMe(reqCtx *gin.Context) {
	user, _ := auth.GetUserFromContext(reqCtx)
	reqCtx.JSON(http.StatusOK, GetMeResponse{
		Object: "me",
		ID:     user.PublicID,
		Email:  user.Email,
		Name:   user.Name,
	})
}

// @Summary Local credential login
// @Description Authenticates an administrator using email and password.
// @Tags Authentication API
// @Accept json
// @Produce json
// @Param request body LocalLoginRequest true "Local login credentials"
// @Success 200 {object} AccessTokenResponse "Successfully authenticated"
// @Failure 400 {object} responses.ErrorResponse "Invalid request payload"
// @Failure 401 {object} responses.ErrorResponse "Invalid credentials"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /v1/auth/local/login [post]
func (authRoute *AuthRoute) LocalLogin(reqCtx *gin.Context) {
	var request LocalLoginRequest
	if err := reqCtx.ShouldBindJSON(&request); err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "f52b5690-620a-11ef-9a5d-5796d531fffe",
			Error: "invalid credentials payload",
		})
		return
	}

	ctx := reqCtx.Request.Context()
	userEntity, err := authRoute.authService.AuthenticateLocalUser(ctx, request.Email, request.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			reqCtx.AbortWithStatusJSON(http.StatusUnauthorized, responses.ErrorResponse{
				Code:  "f52b5a48-620a-11ef-9f49-b3080494c57d",
				Error: "invalid email or password",
			})
			return
		}
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "f52b5bd0-620a-11ef-9520-a7ccb316f52c",
			Error: err.Error(),
		})
		return
	}

	accessTokenExp := time.Now().Add(auth.AccessTokenExpirationDuration)
	accessTokenString, err := auth.CreateJwtSignedString(auth.UserClaim{
		Email: userEntity.Email,
		Name:  userEntity.Name,
		ID:    userEntity.PublicID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExp),
			Subject:   userEntity.Email,
		},
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "f52b5cf0-620a-11ef-abbf-13ac92c1656a",
			Error: err.Error(),
		})
		return
	}

	refreshTokenExp := time.Now().Add(auth.RefreshTokenExpirationDuration)
	refreshTokenString, err := auth.CreateJwtSignedString(auth.UserClaim{
		Email: userEntity.Email,
		Name:  userEntity.Name,
		ID:    userEntity.PublicID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExp),
			Subject:   userEntity.Email,
		},
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "f52b5df8-620a-11ef-8a9a-0fa1e5ac5234",
			Error: err.Error(),
		})
		return
	}

	http.SetCookie(reqCtx.Writer,
		responses.NewCookieWithSecurity(
			auth.RefreshTokenKey,
			refreshTokenString,
			refreshTokenExp,
		),
	)

	reqCtx.JSON(http.StatusOK, &AccessTokenResponse{
		Object:      AccessTokenResponseObjectTypeObject,
		AccessToken: accessTokenString,
		ExpiresIn:   int(time.Until(accessTokenExp).Seconds()),
	})
}

// @Summary Refresh an access token
// @Description Use a valid refresh token to obtain a new access token. The refresh token is typically sent in a cookie.
// @Tags Authentication API
// @Accept json
// @Produce json
// @Success 200 {object} nil "Successfully logout"
// @Failure 400 {object} responses.ErrorResponse "Bad Request (e.g., invalid refresh token)"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized (e.g., expired or missing refresh token)"
// @Router /v1/auth/logout [get]
func (authRoute *AuthRoute) Logout(reqCtx *gin.Context) {
	http.SetCookie(reqCtx.Writer, responses.NewCookieWithSecurity(
		auth.RefreshTokenKey,
		"",
		time.Unix(0, 0),
	))
	reqCtx.Status(http.StatusOK)
}

// @Summary Refresh an access token
// @Description Use a valid refresh token to obtain a new access token. The refresh token is typically sent in a cookie.
// @Tags Authentication API
// @Accept json
// @Produce json
// @Success 200 {object} AccessTokenResponse "Successfully refreshed the access token"
// @Failure 400 {object} responses.ErrorResponse "Bad Request (e.g., invalid refresh token)"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized (e.g., expired or missing refresh token)"
// @Router /v1/auth/refresh-token [get]
func (authRoute *AuthRoute) RefreshToken(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	userClaim, ok := auth.GetUserClaimFromRefreshToken(reqCtx)
	if !ok {
		reqCtx.AbortWithStatusJSON(http.StatusUnauthorized, responses.ErrorResponse{
			Code: "c2019018-b71c-4f13-8ac6-854fbd61c9dd",
		})
		return
	}
	if userClaim.ID == "" {
		user, err := authRoute.userService.FindByEmail(ctx, userClaim.Email)
		if err != nil || user == nil {
			reqCtx.AbortWithStatusJSON(http.StatusUnauthorized, responses.ErrorResponse{
				Code: "58174ddb-ef9c-4a3c-a6ad-c880af070518",
			})
			return
		}
		userClaim.ID = user.PublicID
	}

	accessTokenExp := time.Now().Add(auth.AccessTokenExpirationDuration)
	accessTokenString, err := auth.CreateJwtSignedString(auth.UserClaim{
		Email: userClaim.Email,
		Name:  userClaim.Name,
		ID:    userClaim.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExp),
			Subject:   userClaim.Email,
		},
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "79373f8e-d80e-489c-95ba-9e6099ef7539",
			Error: err.Error(),
		})
		return
	}

	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenString, err := auth.CreateJwtSignedString(auth.UserClaim{
		Email: userClaim.Email,
		Name:  userClaim.Name,
		ID:    userClaim.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExp),
			Subject:   userClaim.Email,
		},
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "0e596742-64bb-4904-8429-4c09ce8434b9",
			Error: err.Error(),
		})
		return
	}

	http.SetCookie(reqCtx.Writer,
		responses.NewCookieWithSecurity(
			auth.RefreshTokenKey,
			refreshTokenString,
			refreshTokenExp,
		),
	)

	reqCtx.JSON(http.StatusOK, &AccessTokenResponse{
		AccessTokenResponseObjectTypeObject,
		accessTokenString,
		int(time.Until(accessTokenExp).Seconds()),
	})
}

// @Summary Guest Login
// @Description JWT-base Guest Login.
// @Tags Authentication API
// @Produce json
// @Success 200 {object} AccessTokenResponse "Successfully refreshed the access token"
// @Failure 400 {object} responses.ErrorResponse "Bad Request (e.g., invalid refresh token)"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized (e.g., expired or missing refresh token)"
// @Router /v1/auth/guest-login [post]
func (authRoute *AuthRoute) GuestLogin(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	userClaim, ok := auth.GetUserClaimFromRefreshToken(reqCtx)
	email := ""
	name := ""
	var id string = ""
	if !ok {
		tempId, err := idgen.GenerateSecureID("jan", 12)
		if err != nil {
			reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
				Code: "3cb11e83-98ed-4c4f-8823-73c26f0c2d75",
			})
			return
		}
		user, err := authRoute.authService.RegisterUser(ctx, &user.User{
			Name:    tempId,
			Email:   fmt.Sprintf("%s@guest.jan.ai", tempId),
			Enabled: true,
			IsGuest: true,
		})
		if err != nil {
			reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
				Code: "9576b6ba-fcc6-4bd2-b13a-33d59d6a71f1",
			})
			return
		}
		email = user.Email
		name = user.Name
		id = user.PublicID
	} else {
		email = userClaim.Email
		name = userClaim.Name
		id = userClaim.ID
	}

	accessTokenExp := time.Now().Add(auth.AccessTokenExpirationDuration)
	accessTokenString, err := auth.CreateJwtSignedString(auth.UserClaim{
		Email: email,
		Name:  name,
		ID:    id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExp),
			Subject:   email,
		},
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "79373f8e-d80e-489c-95ba-9e6099ef7539",
			Error: err.Error(),
		})
		return
	}

	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenString, err := auth.CreateJwtSignedString(auth.UserClaim{
		Email: email,
		Name:  name,
		ID:    id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExp),
			Subject:   email,
		},
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "0e596742-64bb-4904-8429-4c09ce8434b9",
			Error: err.Error(),
		})
		return
	}

	http.SetCookie(reqCtx.Writer, responses.NewCookieWithSecurity(
		auth.RefreshTokenKey,
		refreshTokenString,
		refreshTokenExp,
	))

	reqCtx.JSON(http.StatusOK, &AccessTokenResponse{
		AccessTokenResponseObjectTypeObject,
		accessTokenString,
		int(time.Until(accessTokenExp).Seconds()),
	})
}
