// Copyright (c) 2016. See AUTHORS file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package security

import (
	"fmt"
	"github.com/mbrlabs/zion"
	"strings"
)

// GetUser returns a user from extras of the given context or nil if none found
func GetUser(ctx *zion.Context) User {
	extra := ctx.Extra(zion.ExtraUser)
	if extra != nil {
		user, ok := extra.(User)
		if user != nil && ok {
			return user
		}
	}

	return nil
}

// SecurityStrategy
//------------------------------------------------------------------------------------

type SecurityStrategy interface {
	Authenticate() zion.HandlerFunc
	Logout() zion.HandlerFunc
}

// SecurityRule
//------------------------------------------------------------------------------------

type SecurityRule struct {
	pattern            []string
	allowedHTTPMethods map[string]bool
	userRoles          []string
}

// TODO implement
func (r *SecurityRule) doesPatternMatch(ctx *zion.Context) bool {
	parts := strings.Split(strings.Trim(ctx.Path(), "/"), "/")
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

// SecurityRules
//------------------------------------------------------------------------------------

type SecurityRules []SecurityRule

func NewSecurityRule(pattern string, httpMethods []string, userRoles []string) SecurityRule {
	// convert nils to empty slices
	if httpMethods == nil {
		httpMethods = make([]string, 0)
	}
	if userRoles == nil {
		userRoles = make([]string, 0)
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

func (r SecurityRules) IsAllowed(user User, ctx *zion.Context) bool {
	for _, rule := range r {
		if rule.doesPatternMatch(ctx) {
			// if the user is nil and this route is protected, return false
			if user == nil {
				fmt.Println("[SECURITY] user == nil")
				return false
			}

			// check http method
			if len(rule.allowedHTTPMethods) > 0 {
				if _, ok := rule.allowedHTTPMethods[ctx.Method()]; !ok {
					fmt.Println("[SECURITY] _, ok := rule.allowedHTTPMethods[ctx.Request.Method]; !ok")
					return false
				}
			}

			// check user role
			if len(rule.userRoles) > 0 {
				hasRight := false
				for _, role := range rule.userRoles {
					if user.HasRole(role) {
						hasRight = true
						break
					}
				}
				if !hasRight {
					return false
				}
			}

			fmt.Println("[SECURITY] passed protected area")
			return true
		}
	}

	// no rule found for this path
	fmt.Println("[SECURITY] accessed unprotected area")
	return true
}

// SecurityMiddleware
//------------------------------------------------------------------------------------

type SecurityMiddleware interface {
	zion.Middleware
	AddRule(rule SecurityRule)
}
