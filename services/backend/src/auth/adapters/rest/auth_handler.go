package rest

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"backend/src/api"
	"backend/src/auth/domain"
	"backend/src/auth/ports/input"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	sessionCookie    = "sigta_session"
	loginStateCookie = "sigta_login_state"
	loginStateMaxAge = 10 * time.Minute
	// maxReturnTo keeps the cookie of the pending login small: no page of the
	// application has a longer path.
	maxReturnTo = 512
)

type AuthHandler struct {
	auth          input.Authenticator
	memberships   input.Memberships
	afterLoginURL string
	loginURL      string
	secureCookies bool
}

func NewAuthHandler(auth input.Authenticator, memberships input.Memberships, afterLoginURL, loginURL string, secureCookies bool) *AuthHandler {
	return &AuthHandler{auth: auth, memberships: memberships, afterLoginURL: afterLoginURL, loginURL: loginURL, secureCookies: secureCookies}
}

func (h *AuthHandler) BeginLogin(ctx context.Context, req api.BeginLoginRequestObject) (api.BeginLoginResponseObject, error) {
	location, pending := h.auth.BeginLogin(valueOf(req.Params.LoginHint))
	returnTo := valueOf(req.Params.ReturnTo)
	if _, ok := appPage(returnTo); !ok {
		returnTo = ""
	}
	cookie := h.cookie(loginStateCookie, encodePendingLogin(pending, returnTo), int(loginStateMaxAge.Seconds()))
	return api.BeginLogin302Response{Headers: api.BeginLogin302ResponseHeaders{Location: &location, SetCookie: &cookie}}, nil
}

// CompleteLogin is a browser navigation, so it always redirects: to the app
// with a session, or to the login page with the reason it failed.
func (h *AuthHandler) CompleteLogin(ctx context.Context, req api.CompleteLoginRequestObject) (api.CompleteLoginResponseObject, error) {
	pending, returnTo := decodePendingLogin(valueOf(req.Params.SigtaLoginState))
	token, err := h.completeLogin(ctx, req.Params, pending)

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
	location := h.afterLogin(returnTo)
	redirect := api.CompleteLogin302Response{Headers: api.CompleteLogin302ResponseHeaders{Location: &location}}
	return withCookies{response: redirect, cookies: []string{h.sessionCookie(token), expiredState}}, nil
}

func (h *AuthHandler) completeLogin(ctx context.Context, params api.CompleteLoginParams, pending domain.PendingLogin) (string, error) {
	// The provider answers with error instead of code when the person cancels
	// or it refuses the login.
	code := valueOf(params.Code)
	if params.Error != nil || code == "" {
		return "", domain.ErrLoginRejected
	}
	return h.auth.CompleteLogin(ctx, code, params.State, pending)
}

// afterLogin is where a login leads: back to the page of the application that
// asked for it, or to the home of the application.
func (h *AuthHandler) afterLogin(returnTo string) string {
	home, err := url.Parse(h.afterLoginURL)
	page, ok := appPage(returnTo)
	if err != nil || !ok {
		return h.afterLoginURL
	}
	return home.ResolveReference(page).String()
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
	if input.Unauthenticated(err) {
		return api.GetCurrentUser401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	access, err := h.memberships.Access(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return api.GetCurrentUser200JSONResponse{
		Id:            user.ID,
		Email:         openapi_types.Email(user.Email),
		Name:          user.Name,
		Staff:         access.IsStaff(),
		Administrator: access.Administrator,
	}, nil
}

// sessionCookie has no Max-Age: the session ends on the server after the idle
// timeout, and in the browser when it closes.
func (h *AuthHandler) sessionCookie(token string) string {
	return h.cookie(sessionCookie, token, 0)
}

func (h *AuthHandler) expiredCookie(name string) string {
	return h.cookie(name, "", -1)
}

func (h *AuthHandler) cookie(name, value string, maxAge int) string {
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

func unauthorized() api.UnauthorizedJSONResponse {
	return api.UnauthorizedJSONResponse{Message: messageSignIn}
}

func valueOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// appPage reads returnTo as a page of the application, as the application
// writes it: a path with a single slash in front and no backslash, which
// browsers take for a slash. The page is relative to the home of the
// application, so whatever it says, the login cannot lead to another site.
func appPage(returnTo string) (*url.URL, bool) {
	if len(returnTo) > maxReturnTo || !strings.HasPrefix(returnTo, "/") || strings.HasPrefix(returnTo, "//") || strings.Contains(returnTo, `\`) {
		return nil, false
	}
	page, err := url.Parse("." + returnTo)
	return page, err == nil
}

// The pending login travels in an HttpOnly cookie between leaving for the
// provider and coming back, with the page to return to, if there is one, at
// the end. Its values are base64url, so a dot separates them.
func encodePendingLogin(p domain.PendingLogin, returnTo string) string {
	parts := []string{p.State, p.Nonce, p.CodeVerifier}
	if returnTo != "" {
		parts = append(parts, base64.RawURLEncoding.EncodeToString([]byte(returnTo)))
	}
	return strings.Join(parts, ".")
}

func decodePendingLogin(cookie string) (domain.PendingLogin, string) {
	parts := strings.Split(cookie, ".")
	if len(parts) != 3 && len(parts) != 4 {
		return domain.PendingLogin{}, ""
	}
	pending := domain.PendingLogin{State: parts[0], Nonce: parts[1], CodeVerifier: parts[2]}
	if len(parts) == 3 {
		return pending, ""
	}
	returnTo, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return pending, ""
	}
	return pending, string(returnTo)
}
