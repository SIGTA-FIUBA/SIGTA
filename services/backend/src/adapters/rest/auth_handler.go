package rest

import (
	"backend/src/domain"
	ports "backend/src/ports/input"
	out "backend/src/ports/output"
	"context"
	"errors"
	"net/http"
	"strings"

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
	location, pending, err := h.auth.BeginLogin(valueOf(req.Params.LoginHint))
	if err != nil {
		return nil, err
	}
	cookie := h.cookie(loginStateCookie, encodePendingLogin(pending), loginStateMaxAge)
	return BeginLogin302Response{Headers: BeginLogin302ResponseHeaders{Location: &location, SetCookie: &cookie}}, nil
}

func (h *AuthHandler) CompleteLogin(ctx context.Context, req CompleteLoginRequestObject) (CompleteLoginResponseObject, error) {
	pending := decodePendingLogin(valueOf(req.Params.SigtaLoginState))
	token, err := h.auth.CompleteLogin(ctx, req.Params.Code, req.Params.State, pending)

	// A pending login is good for one attempt, whatever its outcome.
	expiredState := h.cookie(loginStateCookie, "", -1)
	if message := loginFailure(err); message != "" {
		return withCookies{CompleteLogin401JSONResponse{unauthorized(message)}, []string{expiredState}}, nil
	}
	if err != nil {
		return nil, err
	}
	redirect := CompleteLogin302Response{Headers: CompleteLogin302ResponseHeaders{Location: &h.appURL}}
	return withCookies{redirect, []string{h.sessionCookie(token), expiredState}}, nil
}

// withCookies adds cookies to a response. The generated responses hold a single
// Set-Cookie value, and the callback may need to send two.
type withCookies struct {
	response CompleteLoginResponseObject
	cookies  []string
}

func (r withCookies) VisitCompleteLoginResponse(w http.ResponseWriter) error {
	for _, cookie := range r.cookies {
		w.Header().Add("Set-Cookie", cookie)
	}
	return r.response.VisitCompleteLoginResponse(w)
}

// loginFailure is what to tell someone whose login failed for a reason they can
// act on. It is empty for any other outcome.
func loginFailure(err error) string {
	switch {
	case errors.Is(err, out.ErrLoginStateMismatch):
		return "El ingreso no coincide con el que se inició. Volvé a intentarlo."
	case errors.Is(err, out.ErrDomainNotAllowed):
		return "Ingresá con tu cuenta institucional."
	case errors.Is(err, out.ErrLoginRejected):
		return "El ingreso no se pudo completar. Volvé a intentarlo."
	case errors.Is(err, domain.ErrUnknownUser):
		return "No hay una cuenta de SIGTA para esa identidad."
	}
	return ""
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
	if req.Body.Token == "" {
		return RedeemLoginLink400JSONResponse{BadRequestJSONResponse{Message: invalidRequestMessage}}, nil
	}
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

// The pending login travels in an HttpOnly cookie between leaving for the
// provider and coming back. Its values are base64url, so a dot separates them.
func encodePendingLogin(p out.PendingLogin) string {
	return strings.Join([]string{p.State, p.Nonce, p.CodeVerifier}, ".")
}

func decodePendingLogin(cookie string) out.PendingLogin {
	parts := strings.Split(cookie, ".")
	if len(parts) != 3 {
		return out.PendingLogin{}
	}
	return out.PendingLogin{State: parts[0], Nonce: parts[1], CodeVerifier: parts[2]}
}
