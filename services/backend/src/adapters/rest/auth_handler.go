package rest

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"domain"
	"encoding/base64"
	"errors"
	ports "input"
	"net/http"
	out "output"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	sessionCookie    = "sigta_session"
	loginStateCookie = "sigta_login_state"
	loginStateMaxAge = 600
)

type AuthHandler struct {
	auth ports.Authenticator
	// appURL is where the browser lands after logging in.
	appURL string
	// secureCookies must be true everywhere but local development over HTTP.
	secureCookies bool
}

func NewAuthHandler(auth ports.Authenticator, appURL string, secureCookies bool) *AuthHandler {
	return &AuthHandler{auth: auth, appURL: appURL, secureCookies: secureCookies}
}

func (h *AuthHandler) BeginLogin(ctx context.Context, req BeginLoginRequestObject) (BeginLoginResponseObject, error) {
	state, err := newState()
	if err != nil {
		return nil, err
	}
	loginHint := ""
	if req.Params.LoginHint != nil {
		loginHint = *req.Params.LoginHint
	}
	location := h.auth.BeginLogin(state, loginHint)
	cookie := h.cookie(loginStateCookie, state, loginStateMaxAge)
	return BeginLogin302Response{Headers: BeginLogin302ResponseHeaders{Location: &location, SetCookie: &cookie}}, nil
}

func (h *AuthHandler) CompleteLogin(ctx context.Context, req CompleteLoginRequestObject) (CompleteLoginResponseObject, error) {
	expected := req.Params.SigtaLoginState
	if expected == nil || subtle.ConstantTimeCompare([]byte(*expected), []byte(req.Params.State)) != 1 {
		return CompleteLogin401JSONResponse{unauthorized("El ingreso no coincide con el que se inició.")}, nil
	}
	token, err := h.auth.CompleteLogin(ctx, req.Params.Code)
	if errors.Is(err, out.ErrDomainNotAllowed) {
		return CompleteLogin401JSONResponse{unauthorized("Ingresá con tu cuenta institucional.")}, nil
	}
	if errors.Is(err, out.ErrLoginRejected) || errors.Is(err, domain.ErrUnknownUser) {
		return CompleteLogin401JSONResponse{unauthorized("No hay una cuenta de SIGTA para esa identidad.")}, nil
	}
	if err != nil {
		return nil, err
	}
	cookie := h.sessionCookie(token)
	return CompleteLogin302Response{Headers: CompleteLogin302ResponseHeaders{Location: &h.appURL, SetCookie: &cookie}}, nil
}

func (h *AuthHandler) RequestLoginLink(ctx context.Context, req RequestLoginLinkRequestObject) (RequestLoginLinkResponseObject, error) {
	err := h.auth.RequestLoginLink(ctx, string(req.Body.Email))
	if errors.Is(err, out.ErrDomainNotAllowed) {
		return RequestLoginLink400JSONResponse{BadRequestJSONResponse{Message: "Usá tu correo institucional."}}, nil
	}
	if err != nil {
		return nil, err
	}
	return RequestLoginLink202Response{}, nil
}

func (h *AuthHandler) RedeemLoginLink(ctx context.Context, req RedeemLoginLinkRequestObject) (RedeemLoginLinkResponseObject, error) {
	token, err := h.auth.RedeemLoginLink(ctx, req.Body.Token)
	if errors.Is(err, out.ErrInvalidLoginLink) || errors.Is(err, domain.ErrUnknownUser) {
		return RedeemLoginLink401JSONResponse{unauthorized("El enlace venció o ya se usó. Pedí uno nuevo.")}, nil
	}
	if err != nil {
		return nil, err
	}
	cookie := h.sessionCookie(token)
	return RedeemLoginLink204Response{Headers: RedeemLoginLink204ResponseHeaders{SetCookie: &cookie}}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req LogoutRequestObject) (LogoutResponseObject, error) {
	if err := h.auth.Logout(ctx, valueOf(req.Params.SigtaSession)); err != nil {
		return nil, err
	}
	cookie := h.cookie(sessionCookie, "", -1)
	return Logout204Response{Headers: Logout204ResponseHeaders{SetCookie: &cookie}}, nil
}

func (h *AuthHandler) GetCurrentUser(ctx context.Context, req GetCurrentUserRequestObject) (GetCurrentUserResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if isUnauthenticated(err) {
		return GetCurrentUser401JSONResponse{unauthorized("Iniciá sesión para continuar.")}, nil
	}
	if err != nil {
		return nil, err
	}
	return GetCurrentUser200JSONResponse{
		Id:    user.ID,
		Email: openapi_types.Email(user.Email),
		Name:  user.Name,
		Role:  Role(user.Role),
	}, nil
}

// sessionCookie has no Max-Age: the session ends on the server after the idle
// timeout, and in the browser when it closes.
func (h *AuthHandler) sessionCookie(token string) string {
	return h.cookie(sessionCookie, token, 0)
}

func (h *AuthHandler) cookie(name, value string, maxAge int) string {
	c := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
	return c.String()
}

func isUnauthenticated(err error) bool {
	return errors.Is(err, out.ErrSessionNotFound) ||
		errors.Is(err, out.ErrSessionExpired) ||
		errors.Is(err, domain.ErrUnknownUser)
}

func unauthorized(message string) UnauthorizedJSONResponse {
	return UnauthorizedJSONResponse{Message: message}
}

func valueOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func newState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
