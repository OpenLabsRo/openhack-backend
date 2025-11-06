// Package backend provides top-level metadata for the OpenHack API.
//
// @title OpenHack Backend API
// @version 25.11.06.2
// @description Backend API for OpenHack handling participant accounts, teams, feature flags, and superuser check-in tooling.
// @BasePath /
// @securityDefinitions.apikey AccountAuth
// @in header
// @name Authorization
// @description Provide the participant bearer token as `Bearer <token>`.
// @securityDefinitions.apikey SuperUserAuth
// @in header
// @name Authorization
// @description Provide the superuser bearer token as `Bearer <token>`.
package backend

