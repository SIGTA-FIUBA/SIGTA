package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"backend/src/api"
	"backend/src/auth/domain"
	ports "backend/src/auth/ports/input"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	sessionCookie    = "sigta_session"
	loginStateCookie = "sigta_login_state"
	loginStateMaxAge = 10 * time.Minute
)

type AuthHandler struct {
	auth        ports.Authenticator
	memberships ports.Memberships
	// appURL is where the browser lands after logging in.
	appURL   string
	loginURL string
	// secureCookies must be true everywhere but local development over HTTP.
	secureCookies bool
}

func NewAuthHandler(auth ports.Authenticator, memberships ports.Memberships, appURL, loginURL string, secureCookies bool) *AuthHandler {
	return &AuthHandler{auth: auth, memberships: memberships, appURL: appURL, loginURL: loginURL, secureCookies: secureCookies}
}

func (h *AuthHandler) BeginLogin(ctx context.Context, req api.BeginLoginRequestObject) (api.BeginLoginResponseObject, error) {
	location, pending, err := h.auth.BeginLogin(valueOf(req.Params.LoginHint))
	if err != nil {
		return nil, err
	}
	cookie := h.cookie(loginStateCookie, encodePendingLogin(pending), int(loginStateMaxAge.Seconds()))
	return api.BeginLogin302Response{Headers: api.BeginLogin302ResponseHeaders{Location: &location, SetCookie: &cookie}}, nil
}

// CompleteLogin is a browser navigation, so it always redirects: to the app
// with a session, or to the login page with the reason it failed.
func (h *AuthHandler) CompleteLogin(ctx context.Context, req api.CompleteLoginRequestObject) (api.CompleteLoginResponseObject, error) {
	token, err := h.completeLogin(ctx, req.Params)

	// A pending login is good for one attempt, whatever its outcome.
	expiredState := h.expiredCookie(loginStateCookie)
	if reason, failed := loginFailure(err); failed {
		location := h.loginURL + "?" + url.Values{"error": {string(reason)}}.Encode()
		redirect := api.CompleteLogin302Response{Headers: api.CompleteLogin302ResponseHeaders{Location: &location}}
		return withCookies{response: redirect, cookies: []string{expiredState}}, nil
	}
	if err != nil {
		return nil, err
	}
	redirect := api.CompleteLogin302Response{Headers: api.CompleteLogin302ResponseHeaders{Location: &h.appURL}}
	return withCookies{response: redirect, cookies: []string{h.sessionCookie(token), expiredState}}, nil
}

func (h *AuthHandler) completeLogin(ctx context.Context, params api.CompleteLoginParams) (string, error) {
	// The provider answers with error instead of code when the person cancels
	// or it refuses the login.
	code := valueOf(params.Code)
	if params.Error != nil || code == "" {
		return "", domain.ErrLoginRejected
	}
	pending := decodePendingLogin(valueOf(params.SigtaLoginState))
	return h.auth.CompleteLogin(ctx, code, params.State, pending)
}

// withCookies adds cookies to a response. The generated responses hold a single
// Set-Cookie value, and the callback may need to send two.
type withCookies struct {
	response api.CompleteLoginResponseObject
	cookies  []string
}

func (r withCookies) VisitCompleteLoginResponse(w http.ResponseWriter) error {
	for _, cookie := range r.cookies {
		w.Header().Add("Set-Cookie", cookie)
	}
	return r.response.VisitCompleteLoginResponse(w)
}

// loginFailure is the reason a login failed, for the failures the person can
// act on. The login page shows its own text for each reason. The domain goes
// before the rejection: a provider that refuses an account outside the domain
// says so, and the application wraps that in ErrLoginRejected.
func loginFailure(err error) (api.LoginError, bool) {
	switch {
	case errors.Is(err, domain.ErrLoginStateMismatch):
		return api.StateMismatch, true
	case errors.Is(err, domain.ErrDomainNotAllowed):
		return api.DomainNotAllowed, true
	case errors.Is(err, domain.ErrLoginRejected):
		return api.Rejected, true
	}
	return "", false
}

func (h *AuthHandler) Logout(ctx context.Context, req api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	if err := h.auth.Logout(ctx, valueOf(req.Params.SigtaSession)); err != nil {
		return nil, err
	}
	cookie := h.expiredCookie(sessionCookie)
	return api.Logout204Response{Headers: api.Logout204ResponseHeaders{SetCookie: &cookie}}, nil
}

func (h *AuthHandler) GetCurrentUser(ctx context.Context, req api.GetCurrentUserRequestObject) (api.GetCurrentUserResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if isUnauthenticated(err) {
		return api.GetCurrentUser401JSONResponse{UnauthorizedJSONResponse: unauthorized("Iniciá sesión para continuar.")}, nil
	}
	if err != nil {
		return nil, err
	}
	permissions, err := h.memberships.Permissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return api.GetCurrentUser200JSONResponse{
		Id:            user.ID,
		Email:         openapi_types.Email(user.Email),
		Name:          user.Name,
		Role:          api.Role(permissions.Role()),
		Administrator: permissions.Administrator,
	}, nil
}

// sessionCookie has no Max-Age: the session ends on the server after the idle
// timeout, and in the browser when it closes.
func (h *AuthHandler) sessionCookie(token string) string {
	return h.cookie(sessionCookie, token, 0)
}

// expiredCookie tells the browser to delete the cookie.
func (h *AuthHandler) expiredCookie(name string) string {
	return h.cookie(name, "", -1)
}

func (h *AuthHandler) cookie(name, value string, maxAge int) string {
	// Secure is off only for local development over plain HTTP.
	c := http.Cookie{ //nolint:gosec // G124: Secure comes from configuration.
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
	return errors.Is(err, domain.ErrSessionNotFound) ||
		errors.Is(err, domain.ErrSessionExpired) ||
		errors.Is(err, domain.ErrUnknownUser)
}

func unauthorized(message string) api.UnauthorizedJSONResponse {
	return api.UnauthorizedJSONResponse{Message: message}
}

func valueOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// The pending login travels in an HttpOnly cookie between leaving for the
// provider and coming back. Its values are base64url, so a dot separates them.
func encodePendingLogin(p domain.PendingLogin) string {
	return strings.Join([]string{p.State, p.Nonce, p.CodeVerifier}, ".")
}

func decodePendingLogin(cookie string) domain.PendingLogin {
	parts := strings.Split(cookie, ".")
	if len(parts) != 3 {
		return domain.PendingLogin{}
	}
	return domain.PendingLogin{State: parts[0], Nonce: parts[1], CodeVerifier: parts[2]}
}
