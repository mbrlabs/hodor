package security

import (
	"fmt"
	"github.com/mbrlabs/hodor"
	"net/http"
	"strings"
)

type SecurityStrategy interface {
	Authenticate() hodor.HandlerFunc
}

// ============================================================================
// 							Local strategy + middleware
// ============================================================================

type LocalSecurityStrategy struct {
	userStore       UserStore
	sessionStore    SessionStore
	successRedirect string
	failureRedirect string
	loginNameField  string
	passwordField   string
}

func NewLocalSecurityStrategy(userStore UserStore, sessionStore SessionStore) *LocalSecurityStrategy {
	return &LocalSecurityStrategy{
		userStore:    userStore,
		sessionStore: sessionStore,
	}
}

func (ls *LocalSecurityStrategy) SetUserStore(store UserStore) {
	ls.userStore = store
}

func (ls *LocalSecurityStrategy) SetSessionStore(store SessionStore) {
	ls.sessionStore = store
}

func (ls *LocalSecurityStrategy) SetRedirects(successRedirect string, failureRedirect string) {
	ls.failureRedirect = failureRedirect
	ls.successRedirect = successRedirect
}

func (ls *LocalSecurityStrategy) SetPostParameterFields(loginNameField string, passwordField string) {
	ls.loginNameField = loginNameField
	ls.passwordField = passwordField
}

func (ls *LocalSecurityStrategy) Authenticate() hodor.HandlerFunc {
	return func(ctx *hodor.Context) {
		login := ctx.Request.FormValue(ls.loginNameField)
		password := ctx.Request.FormValue(ls.passwordField)

		// handle empty input
		if len(login) == 0 || len(password) == 0 {
			http.Redirect(ctx.Writer, ctx.Request, ls.failureRedirect, http.StatusOK)
			return
		}

		// get user
		user := ls.userStore.GetUserByLogin(login)
		if user == nil {
			http.Redirect(ctx.Writer, ctx.Request, ls.failureRedirect, http.StatusOK)
			return
		}

		// authenticate user
		if ls.userStore.Authenticate(user, password) {
			// create new session
			session := NewSession(user)
			err := ls.sessionStore.Save(session)
			if err == nil {
				// set cockie
				cookie := &http.Cookie{
					Name:    sessionCookieName,
					Value:   session.ID,
					Expires: session.Expire,
				}
				http.SetCookie(ctx.Writer, cookie)
				// redirect to succcess page
				http.Redirect(ctx.Writer, ctx.Request, ls.successRedirect, http.StatusOK)
				return
			}
		}

		// redirect to error page
		http.Redirect(ctx.Writer, ctx.Request, ls.failureRedirect, http.StatusOK)
	}
}

type SecurityRule struct {
	pattern            []string
	allowedHTTPMethods map[string]bool
	userRoles          []hodor.UserRole
}

type SecurityRules []SecurityRule

func NewSecurityRule(pattern string, httpMethods []string, userRoles []hodor.UserRole) SecurityRule {
	// convert nils to empty slices
	if httpMethods == nil {
		httpMethods = make([]string, 0)
	}
	if userRoles == nil {
		userRoles = make([]hodor.UserRole, 0)
	}

	// create a set of http methods
	methods := make(map[string]bool)
	for _, m := range httpMethods {
		methods[m] = true
	}

	// build rule
	return SecurityRule{
		pattern:            strings.Split(strings.Trim(pattern, "/"), "/"),
		allowedHTTPMethods: methods,
		userRoles:          userRoles,
	}
}

func (r SecurityRules) IsAllowed(user hodor.User, ctx *hodor.Context) bool {
	for _, rule := range r {
		if rule.doesPatternMatch(ctx) {
			// if the user is nil and this route is protected, return false
			if user == nil {
				fmt.Println("[SECURITY] user == nil")
				return false
			}

			// check http method
			if len(rule.allowedHTTPMethods) > 0 {
				if _, ok := rule.allowedHTTPMethods[ctx.Request.Method]; !ok {
					fmt.Println("[SECURITY] _, ok := rule.allowedHTTPMethods[ctx.Request.Method]; !ok")
					return false
				}
			}

			// TODO check user roles
			fmt.Println("[SECURITY] passed protected area")
			return true
		}
	}

	// no rule found for this path
	fmt.Println("[SECURITY] accessed unprotected area")
	return true
}

// TODO implement
func (r *SecurityRule) doesPatternMatch(ctx *hodor.Context) bool {
	parts := strings.Split(strings.Trim(ctx.Request.URL.Path, "/"), "/")
	partsLen := len(parts)
	partsLastIndex := partsLen - 1

	// return true if path does not match
	for i, patternPart := range r.pattern {
		if i >= partsLen {
			return false
		}

		pathPart := parts[i]
		if strings.HasPrefix(patternPart, "*") {
			// wildcard matches the whole rest of the path => break loop & check rest of rule
			return true
		} else if strings.HasPrefix(patternPart, ":") || pathPart == patternPart {
			// if last part is reached we have a match
			if partsLastIndex == i {
				return true
			}
			// otherwise continue with next part
			continue
		} else {
			return false
		}
	}

	return false
}

type SecurityMiddleware interface {
	hodor.Middleware
	AddRule(rule SecurityRule)
}

// LocalSecurityMiddleware #
type LocalSecurityMiddleware struct {
	userStore    UserStore
	sessionStore SessionStore
	rules        SecurityRules
}

func NewLocalSecurityMiddleware(userStore UserStore, sessionStore SessionStore) *LocalSecurityMiddleware {
	return &LocalSecurityMiddleware{
		userStore:    userStore,
		sessionStore: sessionStore,
	}
}

func (ls *LocalSecurityMiddleware) SetUserStore(store UserStore) {
	ls.userStore = store
}

func (ls *LocalSecurityMiddleware) SetSessionStore(store SessionStore) {
	ls.sessionStore = store
}

func (ls *LocalSecurityMiddleware) AddRule(rule SecurityRule) {
	ls.rules = append(ls.rules, rule)
}

func (sm *LocalSecurityMiddleware) Execute(ctx *hodor.Context) bool {
	// get cookie from request header
	cookie, err := ctx.Request.Cookie(sessionCookieName)
	if err != nil {
		http.NotFound(ctx.Writer, ctx.Request)
		return false
	}

	// get session based on session key in cookie
	session := sm.sessionStore.Find(cookie.Value)
	var user hodor.User

	// get user by userID stored in session
	if session != nil {
		user = sm.userStore.GetUserByID(session.UserID)
	}

	// go through all security rules
	if sm.rules.IsAllowed(user, ctx) {
		ctx.User = user
		return true
	}

	http.NotFound(ctx.Writer, ctx.Request)
	return false
}

func (sm *LocalSecurityMiddleware) Name() string {
	return "LocalSecurityMiddleware"
}
