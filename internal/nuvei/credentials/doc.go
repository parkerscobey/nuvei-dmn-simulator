// Package credentials verifies Nuvei merchant credentials before send operations.
//
// Verification uses Nuvei /getSessionToken, which authenticates merchant details
// and returns a session token without using /payment or opening an order.
package credentials
